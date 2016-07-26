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

	switch ev.Type {
	case nicolive.EventTypeGot:
		content, _ = json.Marshal(ev.Content.(nicolive.Comment))
	default:
		Logger.Println(ev.String())
	}
	der.cv.Evch <- &Message{
		Domain:  DomainNagome,
		Func:    FuncComment,
		Command: CommCommentAdd,
		Content: content,
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
	var wg sync.WaitGroup

	numProcSendEvent := 5
	wg.Add(numProcSendEvent)
	for i := 0; i < numProcSendEvent; i++ {
		go cv.sendPluginEvent(i, &wg)
	}

	wg.Add(len(cv.Pgns))
	for i := range cv.Pgns {
		go cv.flushPluginIO(i, &wg)
	}

	wg.Add(len(cv.Pgns))
	for i, pg := range cv.Pgns {
		Logger.Println(pg.Name)
		go cv.readPluginMes(i, &wg)
	}

	wg.Wait()

	return
}
