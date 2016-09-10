package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/diginatu/nagome/nicolive"
)

// A CommentViewer is a pair of an Account and a LiveWaku.
type CommentViewer struct {
	Ac      *nicolive.Account
	Lw      *nicolive.LiveWaku
	Cmm     *nicolive.CommentConnection
	Pgns    []*plugin
	TCPPort string
	Evch    chan *Message
	Quit    chan struct{}
	wg      sync.WaitGroup
}

// Run run the CommentViewer and start connecting plugins
func (cv *CommentViewer) Run() {
	defer cv.Cmm.Disconnect()

	cv.loadPlugins()

	cv.wg.Add(1)
	go pluginTCPServer(cv)

	cv.wg.Add(1)
	go sendPluginEvent(cv)

	cv.wg.Wait()

	return
}

// AddPlugin adds new plugin to Pgns
func (cv *CommentViewer) AddPlugin(p *plugin) {
	cv.Pgns = append(cv.Pgns, p)
	p.Init(len(cv.Pgns))
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
			p := new(plugin)
			pPath := filepath.Join(psPath, d.Name())
			err = p.loadPlugin(filepath.Join(pPath, "plugin.yml"))
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

// ProceedNicoEvent will receive events and emits it.
func (cv *CommentViewer) ProceedNicoEvent(ev *nicolive.Event) {
	var con []byte
	var dom, com string

	switch ev.Type {
	case nicolive.EventTypeGot:
		cm, _ := ev.Content.(nicolive.Comment)
		ct := CtCommentGot{
			No:            cm.No,
			Date:          cm.Date,
			UserID:        cm.UserID,
			Raw:           cm.Comment,
			IsPremium:     cm.IsPremium,
			IsBroadcaster: cm.IsCommand,
			IsStaff:       cm.IsStaff,
			IsAnonymity:   cm.IsAnonymity,
			Score:         cm.Score,
		}
		if cm.IsCommand {
			ct.UserName = "Broadcaster"
		} else {
			ct.UserName = ""
		}

		ct.Comment = strings.Replace(cm.Comment, "\n", "<br>", -1)

		dom = DomainComment
		com = CommCommentGot
		con, _ = json.Marshal(ct)

	case nicolive.EventTypeOpen:
		dom = DomainNagome
		com = CommNagomeOpen

	case nicolive.EventTypeClose:
		dom = DomainNagome
		com = CommNagomeClose

	case nicolive.EventTypeHeartBeatGot:
		hb := ev.Content.(nicolive.HeartbeatValue)
		ct := CtNagomeBroadInfo{
			WatchCount:   hb.WatchCount,
			CommentCount: hb.CommentCount,
		}
		dom = DomainNagome
		com = CommNagomeBroadInfo
		con, _ = json.Marshal(ct)

	case nicolive.EventTypeSend:
		dom = DomainNagome
		com = CommNagomeSend

	case nicolive.EventTypeErr:
		log.Println(ev)
		return
	default:
		log.Println(ev)
		return
	}

	cv.Evch <- &Message{
		Domain:  dom,
		Command: com,
		Content: con,
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
	}
	cv.Evch <- t
}
