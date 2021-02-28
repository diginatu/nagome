package nicolive

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/diginatu/nagome/api"
	"github.com/diginatu/nagome/services/utils"
)

const (
	userNameAPITimesAMinute = 6
)

// A Viewer manages nicolive connection, account and live waku and provides functions for comment viewer.
type Viewer struct {
	Ac   *Account
	Lw   *LiveWaku
	Cmm  *CommentConnection
	Evch chan<- *api.Message

	userDB              *UserDB
	userNameAPITimes    int
	userNameAPIFastTime time.Time
	settings            Settings
	savePath            string
	log                 *log.Logger
}

func NewViewer(savePath string, settings Settings, evch chan<- *api.Message, log *log.Logger) (*Viewer, error) {
	udb, err := NewUserDB(filepath.Join(savePath, userDBDirName))
	if err != nil {
		return nil, err
	}

	v := &Viewer{
		savePath: savePath,
		userDB:   udb,
		settings: settings,
		Evch:     evch,
		log:      log,
	}
	return v, nil
}

// Disconnect disconnects current comment connection if connected.
func (v *Viewer) Disconnect() error {
	if v.Cmm == nil {
		return nil
	}

	err := v.Cmm.Disconnect()
	if err != nil {
		return err
	}
	v.Cmm = nil
	v.Lw = nil

	return nil
}

func (v *Viewer) Quit() error {
	return v.userDB.Close()
}

// CheckIntervalAndCreateUser returns a pointer to new struct nicolive.User with fetched user information unless it exceed the limitation of API
func (v *Viewer) CheckIntervalAndCreateUser(id string) (*User, error) {
	// reset API limit
	if time.Now().After(v.userNameAPIFastTime.Add(time.Minute)) {
		v.userNameAPIFastTime = time.Now()
		v.userNameAPITimes = userNameAPITimesAMinute
	}
	if v.userNameAPITimes > 0 {
		u, nerr := CreateUser(id, v.Ac)
		if nerr != nil {
			return nil, nerr
		}
		v.userNameAPITimes--
		return u, nil
	}
	return nil, fmt.Errorf("exceed the limit of fetching user name from web page")
}

// ProceedNicoEvent will receive events and emits it.
func (v *Viewer) ProceedNicoEvent(ev *Event) {
	switch ev.Type {
	case EventTypeCommentGot:
		v.proceedComment(ev)

	case EventTypeCommentOpen:
		v.Evch <- api.NewMessageMust(api.DomainUI, api.CommUIClearComments, nil)
		lv := ev.Content.(*LiveWaku)
		v.log.Println(lv)
		ct := api.CtNagomeBroadOpen{
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
		v.Evch <- api.NewMessageMust(api.DomainNagome, api.CommNagomeBroadOpen, ct)

	case EventTypeCommentClose:
		v.Evch <- api.NewMessageMust(api.DomainNagome, api.CommNagomeBroadClose, nil)

	case EventTypeHeartBeatGot:
		hb := ev.Content.(*HeartbeatValue)
		ct := api.CtNagomeBroadInfo{hb.WatchCount, hb.CommentCount}
		v.Evch <- api.NewMessageMust(api.DomainNagome, api.CommNagomeBroadInfo, ct)

	case EventTypeCommentSend:
		v.Evch <- api.NewMessageMust(api.DomainNagome, api.CommNagomeCommentSend, nil)

	case EventTypeCommentErr:
		nerr := ev.Content.(Error)
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, nerr.TypeString(), nerr.Description(), v.Evch, v.log)

	default:
		v.log.Println(ev)
	}
}

func (v *Viewer) proceedComment(ev *Event) {
	cm, _ := ev.Content.(Comment)
	ct := api.CtCommentGot{
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
	u, err := v.userDB.Fetch(cm.UserID)
	if err != nil {
		err, ok := err.(Error)
		if ok && err.Type() == ErrDBUserNotFound {
		} else {
			v.log.Println(err)
		}
	} else {
		ct.UserName = u.Name
		ct.UserThumbnailURL = u.ThumbnailURL
	}

	if cm.IsCommand {
		ct.UserName = "Broadcaster"
	}

	ct.Comment = strings.Replace(cm.Comment, "\n", "<br>", -1)
	v.Evch <- api.NewMessageMust(api.DomainComment, api.CommCommentGot, ct)

	useAPI := v.settings.UserNameGet && cm.Date.After(v.Cmm.ConnectedTm) && !cm.IsAnonymity && !cm.IsCommand
	if ct.UserName == "" && useAPI {
		v.Evch <- api.NewMessageMust(api.DomainQuery, api.CommQueryUserFetch,
			api.CtQueryUserFetch{ID: ct.UserID})
	}
}
