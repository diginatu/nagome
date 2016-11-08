package main

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

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
	var err error
	waitWakeServer := make(chan struct{})

	cv.wg.Add(2)
	go cv.pluginTCPServer(waitWakeServer)
	go sendPluginMessage(cv)

	<-waitWakeServer
	cv.loadPlugins()

	cv.Antn, err = nicolive.ConnectAntenna(context.TODO(), cv.Ac, nil)
	if err != nil {
		log.Println(err)
		cv.CreateEvNewDialog(CtUIDialogTypeWarn, "Antenna error", "Antenna login failed")
	}

	return
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

	close(waitWakeServer)

	for {
		l.SetDeadline(time.Now().Add(time.Second))
		select {
		default:
			conn, err := l.Accept()
			if err != nil {
				nerr, ok := err.(net.Error)
				if ok && nerr.Timeout() && nerr.Temporary() {
					continue
				}
				log.Println(err)
				continue
			}
			cv.wg.Add(1)
			go handleTCPPlugin(conn, cv)
		case <-cv.quit:
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
