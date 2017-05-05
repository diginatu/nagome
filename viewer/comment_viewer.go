package viewer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	Pgns     []*Plugin
	Settings SettingsSlot
	TCPPort  string
	Evch     chan *Message
	quit     chan struct{}
	wg       sync.WaitGroup
	prcdnle  *ProceedNicoliveEvent
	cli      *CLI
}

// NewCommentViewer makes new CommentViewer
func NewCommentViewer(tcpPort string, cli *CLI) *CommentViewer {
	if len(cli.SettingsSlots.Config) == 0 {
		cli.SettingsSlots.Add(NewSettingsSlot())
	}
	cv := &CommentViewer{
		Settings: cli.SettingsSlots.Config[0].Duplicate(),
		TCPPort:  tcpPort,
		Evch:     make(chan *Message, eventBufferSize),
		quit:     make(chan struct{}),
		cli:      cli,
	}
	cv.prcdnle = NewProceedNicoliveEvent(cv)
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
	cv.AntennaDisconnect()
	var err error
	cv.Antn, err = nicolive.ConnectAntenna(context.TODO(), cv.Ac, cv.prcdnle)
	if err != nil {
		cv.cli.log.Println(err)
		cv.CreateEvNewDialog(CtUIDialogTypeWarn, "Antenna error", "Antenna login failed")
	}
}

// Wait waits for quiting after Start().
func (cv *CommentViewer) Wait() {
	defer cv.AntennaDisconnect()
	defer cv.Disconnect()
	cv.wg.Wait()
}

// Plugin returns plugin with given No.
func (cv *CommentViewer) Plugin(n int) (*Plugin, error) {
	if n < 0 || len(cv.Pgns) <= n {
		return nil, fmt.Errorf("invalid plugin No")
	}
	return cv.Pgns[n], nil
}

// PluginName returns name of the plugin with given No.
func (cv *CommentViewer) PluginName(n int) string {
	if n < -1 || len(cv.Pgns) <= n {
		cv.cli.log.Printf("invalid plugin num : %d\n", n)
		return "???"
	}
	if n == -1 {
		return "NagomeInternal"
	}
	return cv.Pgns[n].Name
}

// AddPlugin adds new plugin to Pgns
func (cv *CommentViewer) AddPlugin(p *Plugin) {
	p.No = len(cv.Pgns)
	cv.Pgns = append(cv.Pgns, p)
}

func (cv *CommentViewer) loadPlugins() {
	psPath := filepath.Join(cv.cli.SavePath, pluginDirName)

	ds, err := ioutil.ReadDir(psPath)
	if err != nil {
		cv.cli.log.Println(err)
		return
	}

	for _, d := range ds {
		if d.IsDir() {
			p := newPlugin(cv)
			pPath := filepath.Join(psPath, d.Name())
			err = p.Load(filepath.Join(pPath, "plugin.yml"))
			if err != nil {
				cv.cli.log.Println("failed load plugin : ", d.Name())
				cv.cli.log.Println(err)
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
					cmd.Dir = pPath
					err := cmd.Start()
					if err != nil {
						cv.cli.log.Println(err)
						continue
					}
				}
			case pluginMethodStd:
				cv.wg.Add(1)
				go handleSTDPlugin(p, cv, pPath)
			default:
				cv.cli.log.Printf("invalid method in plugin [%s]\n", p.Name)
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
		cv.cli.log.Panicln(err)
	}
	l, err := net.ListenTCP("tcp", adr)
	if err != nil {
		cv.cli.log.Panicln(err)
	}
	defer func() {
		err := l.Close()
		if err != nil {
			cv.cli.log.Println(err)
		}
	}()

	_, cv.TCPPort, err = net.SplitHostPort(l.Addr().String())
	if err != nil {
		cv.cli.log.Panicln(err)
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
					cv.cli.log.Println(err)
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

	<-cv.quit
}

func (cv *CommentViewer) sendPluginMessage() {
	defer cv.wg.Done()

	for {
	readLoop:
		select {
		case mes := <-cv.Evch:
			// Direct
			if mes.Domain == DomainDirect {
				nicoerr := processDirectMessage(cv, mes)
				if nicoerr != nil {
					cv.cli.log.Printf("plugin message error form [%s] : %s\n", cv.PluginName(mes.prgno), nicoerr)
					cv.cli.log.Println(mes)
				}
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
				if cv.Pgns[i].IsSubscribe(mes.Domain + DomainSuffixFilter) {
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
				cv.cli.log.Println(err)
				cv.cli.log.Println(mes)
				continue
			}

			// regular
			for i := range cv.Pgns {
				if cv.Pgns[i].IsSubscribe(mes.Domain) {
					cv.Pgns[i].Write(jmes)
				}
			}

			nerr := processPluginMessage(cv, mes)
			if nerr != nil {
				cv.cli.log.Printf("Error : message form [%s] %s\n", cv.PluginName(mes.prgno), nerr)
				cv.cli.log.Println(mes)

				nicoerr, ok := nerr.(nicolive.Error)
				if ok {
					cv.ProceedNicoliveError(nicoerr)
				} else {
					cv.cli.log.Panicln(nerr)
				}
			}

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
	cv.cli.log.Printf("[D] %s : %s", title, desc)
	cv.Evch <- NewMessageMust(DomainUI, CommUIDialog, CtUIDialog{typ, title, desc})
}

// Disconnect disconnects current comment connection if connected.
func (cv *CommentViewer) Disconnect() {
	if cv.Cmm == nil {
		return
	}

	err := cv.Cmm.Disconnect()
	if err != nil {
		cv.cli.log.Println(err)
	}
	cv.Cmm = nil
	cv.Lw = nil

	return
}

// AntennaDisconnect disconnects current antenna connection if connected.
func (cv *CommentViewer) AntennaDisconnect() {
	if cv.Antn == nil {
		return
	}

	err := cv.Antn.Disconnect()
	if err != nil {
		cv.cli.log.Println(err)
	}
	cv.Antn = nil

	return
}

// Quit quits the CommentViewer.
func (cv *CommentViewer) Quit() {
	cv.wg.Add(1)
	defer cv.wg.Done()

	close(cv.quit)
	err := cv.prcdnle.userDB.Close()
	if err != nil {
		cv.cli.log.Println(err)
	}
}

// ProceedNicoliveError proceeds Error of nicolive.
func (cv *CommentViewer) ProceedNicoliveError(e nicolive.Error) {
	switch e.No() {
	case nicolive.ErrOther:
	case nicolive.ErrSendComment:
		cv.CreateEvNewDialog(CtUIDialogTypeWarn, e.TypeString(), e.Description())
	case nicolive.ErrConnection:
	case nicolive.ErrNicoLiveOther:
	case nicolive.ErrNotLogin:
	case nicolive.ErrClosed:
	case nicolive.ErrIncorrectAccount:
	default:
		cv.cli.log.Println("Unknown nicolive Error")
	}
	cv.CreateEvNewDialog(CtUIDialogTypeWarn, e.TypeString(), e.Description())
}
