package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/diginatu/nagome/nicolive"
)

// ProceedNicoliveEvent is struct for proceeding nico events from nicolive packeage.
type ProceedNicoliveEvent struct {
	cv *CommentViewer
}

// NewProceedNicoliveEvent makes new ProceedNicoliveEvent and returns it.
func NewProceedNicoliveEvent(cv *CommentViewer) *ProceedNicoliveEvent {
	return &ProceedNicoliveEvent{cv: cv}
}

// ProceedNicoEvent will receive events and emits it.
func (p *ProceedNicoliveEvent) ProceedNicoEvent(ev *nicolive.Event) {
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
		p.cv.Evch <- &Message{
			Domain:  DomainUI,
			Command: CommUIClearComments,
		}

		dom = DomainNagome
		com = CommNagomeBroadOpen

	case nicolive.EventTypeClose:
		dom = DomainNagome
		com = CommNagomeBroadClose

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
		com = CommNagomeCommentSend

	case nicolive.EventTypeErr:
		log.Println(ev)
		return

	default:
		log.Println(ev)
		return
	}

	p.cv.Evch <- &Message{
		Domain:  dom,
		Command: com,
		Content: con,
	}
}
