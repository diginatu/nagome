package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/diginatu/nagome/nicolive"
)

// commentEventEmit will receive comment events and emits commentViewer events.
type commentEventEmit struct {
	cv *commentViewer
}

func (der *commentEventEmit) Proceed(ev *nicolive.Event) {
	var content []byte
	var command string

	switch ev.Type {
	case nicolive.EventTypeGot:
		content, _ = json.Marshal(ev.Content.(nicolive.Comment))
		command = CommCommentAdd
	default:
		log.Println(ev.String())
	}

	if command != "" {
		der.cv.Evch <- &Message{
			Domain:  DomainNagome,
			Func:    FuncComment,
			Command: command,
			Content: content,
		}
	}
}

// A commentViewer is a pair of an Account and a LiveWaku.
type commentViewer struct {
	Ac     *nicolive.Account
	Lw     *nicolive.LiveWaku
	Cmm    *nicolive.CommentConnection
	Pgns   []*plugin
	Evch   chan *Message
	Quit   chan struct{}
	addpgn chan *plugin
}

func (cv *commentViewer) Run() {
	defer cv.Cmm.Disconnect()
	var wg sync.WaitGroup

	wg.Add(1)
	go cv.addPluginRoutine(&wg)

	wg.Add(1)
	go pluginTCPServer(cv, &wg)

	wg.Add(1)
	go sendPluginEvent(cv, &wg)

	wg.Add(len(cv.Pgns))
	for i, pg := range cv.Pgns {
		log.Println(pg.Name)
		go eachPluginRw(cv, i, &wg)
	}

	wg.Wait()

	return
}

func (cv *commentViewer) AddPlugin(p *plugin) {
	cv.addpgn <- p
}

func (cv *commentViewer) addPluginRoutine(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case t := <-cv.addpgn:
			cv.Pgns = append(cv.Pgns, t)
		case <-cv.Quit:
			return
		}
	}
}

func (cv *commentViewer) CreateEvNewDialog(typ, title, desc string) {
	t, err := NewMessage(DomainNagome, FuncUI, CommUIDialog,
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
