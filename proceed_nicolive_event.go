package main

import (
	"encoding/json"
	"log"
	"path"
	"strings"

	"github.com/diginatu/nagome/nicolive"
)

// ProceedNicoliveEvent is struct for proceeding nico events from nicolive packeage.
type ProceedNicoliveEvent struct {
	cv     *CommentViewer
	userDB *nicolive.UserDB
}

// NewProceedNicoliveEvent makes new ProceedNicoliveEvent and returns it.
func NewProceedNicoliveEvent(cv *CommentViewer) *ProceedNicoliveEvent {
	udb, err := nicolive.NewUserDB(path.Join(App.SavePath, userDBFileName))
	if err != nil {
		log.Fatalln(err)
	}
	return &ProceedNicoliveEvent{
		cv:     cv,
		userDB: udb,
	}
}

func (p *ProceedNicoliveEvent) getUserName(id string, useAPI bool) (*nicolive.User, error) {
	u, err := p.userDB.Fetch(id)
	if err != nil {
		return nil, err
	}
	if useAPI {
		if u == nil {
			u, nerr := nicolive.FetchUserInfo(id, p.cv.Ac)
			if nerr != nil {
				return nil, nerr
			}
			err := p.userDB.Store(u)
			if err != nil {
				return nil, err
			}
			return u, nil
		}
	}
	return u, nil
}

func (p *ProceedNicoliveEvent) proceedComment(ev *nicolive.Event) {
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

	// user info
	u, err := p.getUserName(cm.UserID, !cm.IsAnonymity)
	if err != nil {
		log.Println(err)
	}
	if u != nil {
		ct.UserName = u.Name
		ct.UserThumbnailURL = u.ThumbnailURL
	}
	if cm.IsCommand {
		ct.UserName = "Broadcaster"
	}

	ct.Comment = strings.Replace(cm.Comment, "\n", "<br>", -1)

	con, err := json.Marshal(ct)
	if err != nil {
		log.Println(err)
		return
	}

	p.cv.Evch <- &Message{
		Domain:  DomainComment,
		Command: CommCommentGot,
		Content: con,
	}
}

// ProceedNicoEvent will receive events and emits it.
func (p *ProceedNicoliveEvent) ProceedNicoEvent(ev *nicolive.Event) {
	switch ev.Type {
	case nicolive.EventTypeGot:
		p.proceedComment(ev)
	case nicolive.EventTypeOpen:
		p.cv.Evch <- &Message{
			Domain:  DomainUI,
			Command: CommUIClearComments,
		}
		p.cv.Evch <- &Message{
			Domain:  DomainNagome,
			Command: CommNagomeBroadOpen,
		}
	case nicolive.EventTypeClose:
		p.cv.Evch <- &Message{
			Domain:  DomainNagome,
			Command: CommNagomeBroadClose,
		}
	case nicolive.EventTypeHeartBeatGot:
		hb := ev.Content.(nicolive.HeartbeatValue)
		ct := CtNagomeBroadInfo{
			WatchCount:   hb.WatchCount,
			CommentCount: hb.CommentCount,
		}
		con, err := json.Marshal(ct)
		if err != nil {
			log.Println(err)
			return
		}

		p.cv.Evch <- &Message{
			Domain:  DomainNagome,
			Command: CommNagomeBroadInfo,
			Content: con,
		}
	case nicolive.EventTypeSend:
		p.cv.Evch <- &Message{
			Domain:  DomainNagome,
			Command: CommNagomeCommentSend,
		}
	case nicolive.EventTypeErr:
		log.Println(ev)
		return

	default:
		log.Println(ev)
		return
	}

}
