package main

import (
	"encoding/json"
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
		Logger.Println(ev.String())
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
	Ac   *nicolive.Account
	Lw   *nicolive.LiveWaku
	Cmm  *nicolive.CommentConnection
	Pgns []*plugin
	Evch chan *Message
	Quit chan struct{}
}

func (cv *commentViewer) runCommentViewer() {
	defer cv.Cmm.Disconnect()
	var wg sync.WaitGroup

	wg.Add(1)
	go pluginTCPServer(cv, &wg)

	wg.Add(1)
	go sendPluginEvent(cv, &wg)

	wg.Add(len(cv.Pgns))
	for i, pg := range cv.Pgns {
		Logger.Println(pg.Name)
		go eachPluginRw(cv, i, &wg)
	}

	wg.Wait()

	return
}

func (cv *commentViewer) createEvNewDialog(typ, title, desc string) {
	t, err := NewMessage(DomainNagome, FuncUI, CommUIDialog,
		CtUIDialog{
			Type:        typ,
			Title:       title,
			Description: desc,
		})
	if err != nil {
		Logger.Println(err)
	}
	cv.Evch <- t
}
