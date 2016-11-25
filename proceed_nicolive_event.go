package main

import (
	"log"
	"path/filepath"
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
	udb, err := nicolive.NewUserDB(filepath.Join(App.SavePath, userDBFileName))
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
	useAPI := p.cv.Settings.UserNameGet && cm.Date.After(p.cv.Cmm.ConnectedTm) && !cm.IsAnonymity && !cm.IsCommand
	u, err := p.getUserName(cm.UserID, useAPI)
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
	p.cv.Evch <- NewMessageMust(DomainComment, CommCommentGot, ct)
}

// ProceedNicoEvent will receive events and emits it.
func (p *ProceedNicoliveEvent) ProceedNicoEvent(ev *nicolive.Event) {
	switch ev.Type {
	case nicolive.EventTypeGot:
		p.proceedComment(ev)

	case nicolive.EventTypeOpen:
		p.cv.Evch <- NewMessageMust(DomainUI, CommUIClearComments, nil)
		lv := ev.Content.(*nicolive.LiveWaku)
		log.Println(lv)
		ct := CtNagomeBroadOpen{
			BroadID:     lv.BroadID,
			Title:       lv.Stream.Title,
			Description: lv.Stream.Description,
			CommunityID: lv.Stream.CommunityID,
			OwnerID:     lv.Stream.OwnerID,
			OwnerName:   lv.Stream.OwnerName,
			OwnerBroad:  lv.OwnerBroad,
			OpenTime:    lv.Stream.OpenTime,
			StartTime:   lv.Stream.StartTime,
			EndTime:     lv.Stream.EndTime,
		}
		p.cv.Evch <- NewMessageMust(DomainNagome, CommNagomeBroadOpen, ct)

	case nicolive.EventTypeClose:
		p.cv.Evch <- NewMessageMust(DomainNagome, CommNagomeBroadClose, nil)

	case nicolive.EventTypeHeartBeatGot:
		hb := ev.Content.(*nicolive.HeartbeatValue)
		ct := CtNagomeBroadInfo{hb.WatchCount, hb.CommentCount}
		p.cv.Evch <- NewMessageMust(DomainNagome, CommNagomeBroadInfo, ct)

	case nicolive.EventTypeSend:
		p.cv.Evch <- NewMessageMust(DomainNagome, CommNagomeCommentSend, nil)

	case nicolive.EventTypeErr:
		log.Println(ev)

	case nicolive.EventTypeAntennaOpen:
		p.cv.Evch <- NewMessageMust(DomainNagome, CommNagomeAntennaOpen, nil)

	case nicolive.EventTypeAntennaClose:
		p.cv.Evch <- NewMessageMust(DomainNagome, CommNagomeAntennaClose, nil)

	case nicolive.EventTypeAntennaErr:
		log.Println(ev)

	case nicolive.EventTypeAntennaGot:
		ai := ev.Content.(*nicolive.AntennaItem)
		ct := CtAntennaGot{ai.BroadID, ai.CommunityID, ai.UserID}
		p.cv.Evch <- NewMessageMust(DomainAntenna, CommAntennaGot, ct)

		if p.cv.Settings.AutoFollowNextWaku {
			if p.cv.Lw != nil && p.cv.Lw.Stream.CommunityID == ai.CommunityID {
				ct := CtQueryBroadConnect{ai.BroadID}
				log.Println("following to " + ai.BroadID)
				p.cv.Evch <- NewMessageMust(DomainQuery, CommQueryBroadConnect, ct)
			}
		}

	default:
		log.Println(ev)
	}
}
