package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/diginatu/nagome/nicolive"
)

// A CommentViewer is a pair of an Account and a LiveWaku.
type CommentViewer struct {
	Ac       *nicolive.Account
	Lw       *nicolive.LiveWaku
	Cmm      *nicolive.CommentConnection
	Antn     *nicolive.Antenna
	Pgns     []*plugin
	Settings SettingsSlot
	TCPPort  string
	Evch     chan *Message
	quit     chan struct{}
	wg       sync.WaitGroup
	prcdnle  *ProceedNicoliveEvent
}

// NewCommentViewer makes new CommentViewer
func NewCommentViewer(ac *nicolive.Account, tcpPort string) *CommentViewer {
	if len(App.SettingsSlots.Config) == 0 {
		App.SettingsSlots.Add(NewSettingsSlot())
	}
	cv := &CommentViewer{
		Ac:       ac,
		Settings: *App.SettingsSlots.Config[0],
		TCPPort:  tcpPort,
		Evch:     make(chan *Message, eventBufferSize),
		quit:     make(chan struct{}),
	}
	return cv
}

// Start run the CommentViewer and start connecting plugins
func (cv *CommentViewer) Start() {
	waitWakeServer := make(chan struct{})

	cv.wg.Add(2)
	go cv.pluginTCPServer(waitWakeServer)
	go cv.sendPluginMessage()

	<-waitWakeServer
	cv.loadPlugins()

	return
}

// AntennaConnect connects Antenna and start processing.
func (cv *CommentViewer) AntennaConnect() {
	var err error
	cv.Antn, err = nicolive.ConnectAntenna(context.TODO(), cv.Ac, nil)
	if err != nil {
		log.Println(err)
		cv.CreateEvNewDialog(CtUIDialogTypeWarn, "Antenna error", "Antenna login failed")
	}
}

// Wait waits for quiting after Start().
func (cv *CommentViewer) Wait() {
	defer cv.AntennaDisconnect()
	defer cv.Disconnect()
	cv.wg.Wait()
}

// AddPlugin adds new plugin to Pgns
func (cv *CommentViewer) AddPlugin(p *plugin) {
	cv.Pgns = append(cv.Pgns, p)
	p.No = len(cv.Pgns)
}

func (cv *CommentViewer) loadPlugins() {
	psPath := filepath.Join(App.SavePath, pluginDirName)

	ds, err := ioutil.ReadDir(psPath)
	if err != nil {
		log.Println(err)
		return
	}

	for _, d := range ds {
		if d.IsDir() {
			p := newPlugin(cv)
			pPath := filepath.Join(psPath, d.Name())
			err = p.Load(filepath.Join(pPath, "plugin.yml"))
			if err != nil {
				log.Println("failed load plugin : ", d.Name())
				log.Println(err)
				continue
			}

			cv.AddPlugin(p)

			for i := range p.Exec {
				p.Exec[i] = strings.Replace(p.Exec[i], "{{path}}", pPath, -1)
				p.Exec[i] = strings.Replace(p.Exec[i], "{{port}}", cv.TCPPort, -1)
				p.Exec[i] = strings.Replace(p.Exec[i], "{{no}}", strconv.Itoa(p.No), -1)
			}

			switch p.Method {
			case pluginMethodTCP:
				if len(p.Exec) >= 1 {
					cmd := exec.Command(p.Exec[0], p.Exec[1:]...)
					err := cmd.Start()
					if err != nil {
						log.Println(err)
						continue
					}
				}
			case pluginMethodStd:
				cv.wg.Add(1)
				go handleSTDPlugin(p, cv)
			default:
				log.Printf("invalid method in plugin [%s]\n", p.Name)
				continue
			}
		}
	}

	return
}

func (cv *CommentViewer) pluginTCPServer(waitWakeServer chan struct{}) {
	defer cv.wg.Done()

	adr, err := net.ResolveTCPAddr("tcp", ":"+cv.TCPPort)
	if err != nil {
		log.Panicln(err)
	}
	l, err := net.ListenTCP("tcp", adr)
	if err != nil {
		log.Panicln(err)
	}
	defer l.Close()

	_, cv.TCPPort, err = net.SplitHostPort(l.Addr().String())
	if err != nil {
		log.Panicln(err)
	}

	cv.wg.Add(1)
	go func() {
		defer cv.wg.Done()
		for {
			conn, err := l.Accept()
			if err != nil {
				nerr, ok := err.(net.Error)
				if ok && nerr.Temporary() {
					continue
				}
				select {
				default:
					log.Println(err)
					cv.Quit()
				case <-cv.quit:
				}
				return
			}
			cv.wg.Add(1)
			go handleTCPPlugin(conn, cv)
		}
	}()

	close(waitWakeServer)

	select {
	case <-cv.quit:
		return
	}
}

func (cv *CommentViewer) sendPluginMessage() {
	defer cv.wg.Done()

	for {
	readLoop:
		select {
		case mes := <-cv.Evch:
			// Direct
			if mes.Domain == DomainDirectngm {
				cv.Pgns[mes.prgno].WriteMess(mes)
				continue
			}
			if mes.Domain == DomainDirect {
				go func() {
					nicoerr := processDirectMessage(cv, mes)
					if nicoerr != nil {
						log.Printf("plugin message error form [%s] : %s\n", cv.Pgns[mes.prgno].Name, nicoerr)
						log.Println(mes)
					}
				}()
				continue
			}

			// filter

			// Messages from filter plugin will not send same plugin.
			var st int
			if strings.HasSuffix(mes.Domain, DomainSuffixFilter) {
				st = mes.prgno + 1
				mes.Domain = strings.TrimSuffix(mes.Domain, DomainSuffixFilter)
			}
			for i := st; i < len(cv.Pgns); i++ {
				if cv.Pgns[i].Depend(mes.Domain + DomainSuffixFilter) {
					// Add suffix to a message for filter plugin.
					tmes := *mes
					tmes.Domain = mes.Domain + DomainSuffixFilter
					fail := cv.Pgns[i].WriteMess(&tmes)
					if fail {
						continue
					}
					break readLoop
				}
			}

			jmes, err := json.Marshal(mes)
			if err != nil {
				log.Println(err)
				log.Println(mes)
				continue
			}

			// regular
			for i := range cv.Pgns {
				if cv.Pgns[i].Depend(mes.Domain) {
					cv.Pgns[i].Write(jmes)
				}
			}

			go func() {
				nicoerr := processPluginMessage(cv, mes)
				if nicoerr != nil {
					log.Printf("plugin message error form [%s] : %s\n", cv.Pgns[mes.prgno].Name, nicoerr)
					log.Println(mes)
				}
			}()

		case <-cv.quit:
			for _, p := range cv.Pgns {
				p.Close()
			}
			return
		}
	}
}

// CreateEvNewDialog emits new event for ask UI to display dialog.
func (cv *CommentViewer) CreateEvNewDialog(typ, title, desc string) {
	t, err := NewMessage(DomainUI, CommUIDialog,
		CtUIDialog{
			Type:        typ,
			Title:       title,
			Description: desc,
		})
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("[D] %s : %s", title, desc)
	cv.Evch <- t
}

// Disconnect disconnects current comment connection if connected.
func (cv *CommentViewer) Disconnect() {
	if cv.Cmm == nil {
		return
	}

	err := cv.Cmm.Disconnect()
	if err != nil {
		log.Println(err)
	}
	cv.Cmm = nil

	return
}

// AntennaDisconnect disconnects current antenna connection if connected.
func (cv *CommentViewer) AntennaDisconnect() {
	if cv.Antn == nil {
		return
	}

	err := cv.Antn.Disconnect()
	if err != nil {
		log.Println(err)
	}
	cv.Antn = nil

	return
}

// Quit quits the CommentViewer.
func (cv *CommentViewer) Quit() {
	close(cv.quit)
}
