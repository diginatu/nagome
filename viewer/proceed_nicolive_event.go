package viewer

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/diginatu/nagome/nicolive"
)

const (
	userNameAPITimesAMinute = 6
)

// ProceedNicoliveEvent is struct for proceeding nico events from nicolive packeage.
type ProceedNicoliveEvent struct {
	cv                  *CommentViewer
	userDB              *nicolive.UserDB
	userNameAPITimes    int
	userNameAPIFastTime time.Time
}

// NewProceedNicoliveEvent makes new ProceedNicoliveEvent and returns it.
func NewProceedNicoliveEvent(cv *CommentViewer) *ProceedNicoliveEvent {
	udb, err := nicolive.NewUserDB(filepath.Join(cv.cli.SavePath, userDBDirName))
	if err != nil {
		cv.cli.log.Fatalln(err)
	}
	return &ProceedNicoliveEvent{
		cv:     cv,
		userDB: udb,
	}
}

// CheckIntervalAndCreateUser returns a pointer to new struct nicolive.User with fetched user information unless it exceed the limitation of API
func (p *ProceedNicoliveEvent) CheckIntervalAndCreateUser(id string) (*nicolive.User, error) {
	// reset API limit
	if time.Now().After(p.userNameAPIFastTime.Add(time.Minute)) {
		p.userNameAPIFastTime = time.Now()
		p.userNameAPITimes = userNameAPITimesAMinute
	}
	if p.userNameAPITimes > 0 {
		u, nerr := nicolive.CreateUser(id, p.cv.Ac)
		if nerr != nil {
			return nil, nerr
		}
		p.userNameAPITimes--
		return u, nil
	}
	return nil, fmt.Errorf("exceed the limit of fetching user name from web page")
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

	// Get user name from DB
	u, err := p.userDB.Fetch(cm.UserID)
	if err != nil {
		if err, ok := err.(nicolive.Error); ok {
			if err.Type() == nicolive.ErrDBUserNotFound {
			} else {
				p.cv.cli.log.Println(err)
			}
		} else {
			p.cv.cli.log.Println(err)
		}
	} else {
		ct.UserName = u.Name
		ct.UserThumbnailURL = u.ThumbnailURL
	}

	if cm.IsCommand {
		ct.UserName = "Broadcaster"
	}

	ct.Comment = strings.Replace(cm.Comment, "\n", "<br>", -1)
	p.cv.Evch <- NewMessageMust(DomainComment, CommCommentGot, ct)

	useAPI := p.cv.Settings.UserNameGet && cm.Date.After(p.cv.Cmm.ConnectedTm) && !cm.IsAnonymity && !cm.IsCommand
	if ct.UserName == "" && useAPI {
		p.cv.Evch <- NewMessageMust(DomainQuery, CommQueryUserFetch,
			CtQueryUserFetch{ID: ct.UserID})
	}
}

// ProceedNicoEvent will receive events and emits it.
func (p *ProceedNicoliveEvent) ProceedNicoEvent(ev *nicolive.Event) {
	switch ev.Type {
	case nicolive.EventTypeCommentGot:
		p.proceedComment(ev)

	case nicolive.EventTypeCommentOpen:
		p.cv.Evch <- NewMessageMust(DomainUI, CommUIClearComments, nil)
		lv := ev.Content.(*nicolive.LiveWaku)
		p.cv.cli.log.Println(lv)
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

	case nicolive.EventTypeCommentClose:
		p.cv.Evch <- NewMessageMust(DomainNagome, CommNagomeBroadClose, nil)

	case nicolive.EventTypeHeartBeatGot:
		hb := ev.Content.(*nicolive.HeartbeatValue)
		ct := CtNagomeBroadInfo{hb.WatchCount, hb.CommentCount}
		p.cv.Evch <- NewMessageMust(DomainNagome, CommNagomeBroadInfo, ct)

	case nicolive.EventTypeCommentSend:
		p.cv.Evch <- NewMessageMust(DomainNagome, CommNagomeCommentSend, nil)

	case nicolive.EventTypeCommentErr:
		nerr := ev.Content.(nicolive.Error)
		p.cv.EmitEvNewNotification(CtUINotificationTypeWarn, nerr.TypeString(), nerr.Description())

	case nicolive.EventTypeAntennaOpen:
		p.cv.Evch <- NewMessageMust(DomainNagome, CommNagomeAntennaOpen, nil)

	case nicolive.EventTypeAntennaClose:
		p.cv.Evch <- NewMessageMust(DomainNagome, CommNagomeAntennaClose, nil)

	case nicolive.EventTypeAntennaErr:
		p.cv.cli.log.Println(ev)

	case nicolive.EventTypeAntennaGot:
		ai := ev.Content.(*nicolive.AntennaItem)
		ct := CtAntennaGot{ai.BroadID, ai.CommunityID, ai.UserID}
		p.cv.Evch <- NewMessageMust(DomainAntenna, CommAntennaGot, ct)

	default:
		p.cv.cli.log.Println(ev)
	}
}
