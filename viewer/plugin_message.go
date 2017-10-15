package viewer

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

	"github.com/diginatu/nagome/nicolive"
)

// Constant values for processing plugin messages.
const (
	NumFetchInformaionRetry = 3
)

var (
	broadIDRegex = regexp.MustCompile(`(lv|co)\d+`)
)

func processPluginMessage(cv *CommentViewer, m *Message) error {
	switch m.Domain {
	case DomainQuery:
		switch m.Command {
		case CommQueryBroadConnect:
			var err error
			var ct CtQueryBroadConnect
			if err = json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			broadMch := broadIDRegex.FindString(ct.BroadID)
			if broadMch == "" {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "invalid BroadID", "no valid BroadID found in the ID text")
				return nicolive.MakeError(nicolive.ErrOther, "no valid BroadID found in the ID text")
			}

			lw := &nicolive.LiveWaku{Account: cv.Ac, BroadID: broadMch}

			err = lw.FetchInformation()
			if err != nil {
				nerr, ok := err.(nicolive.Error)
				if ok {
					if nerr.Type() == nicolive.ErrNetwork {
						ct.RetryN++
						cv.cli.log.Printf("Failed to connect to %s.\n", ct.BroadID)
						cv.cli.log.Printf("FetchInformation : %s\n", nerr.Error())
						if ct.RetryN >= NumFetchInformaionRetry {
							cv.cli.log.Println("Reached the limit of retrying.")
							return nerr
						}
						cv.cli.log.Println("Retrying...")
						go func() {
							<-time.After(time.Second)
							cv.Evch <- NewMessageMust(DomainQuery, CommQueryBroadConnect, ct)
						}()
					}
				}
				return err
			}

			cv.Disconnect()

			cv.Lw = lw
			cv.Cmm, err = nicolive.CommentConnect(context.TODO(), *cv.Lw, cv.prcdnle)
			if err != nil {
				return err
			}

			if lw.IsUserOwner() {
				ps, err := nicolive.PublishStatus(lw.BroadID, cv.Ac)
				if err != nil {
					return err
				}
				cv.Lw.OwnerCommentToken = ps.Token
			}

			cv.cli.log.Println("connected")

		case CommQueryBroadDisconnect:
			cv.Disconnect()

		case CommQueryBroadSendComment:
			if cv.Cmm == nil {
				return nicolive.MakeError(nicolive.ErrSendComment, "not connected to live")
			}
			if cv.Lw == nil {
				return nicolive.MakeError(nicolive.ErrSendComment, "Error : cv.Lw is nil")
			}

			var ct CtQueryBroadSendComment
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			isowner := false
			if ct.Type == "" {
				isowner = cv.Settings.OwnerComment
			} else {
				isowner = ct.Type == CtQueryBroadSendCommentTypeOwner
			}
			if !cv.Lw.IsUserOwner() {
				isowner = false
			}

			if isowner {
				rq := nicolive.CommentOwnerRequest{
					Text:        ct.Text,
					Color:       "",
					IsPermanent: false,
					UserName:    "",
				}
				err := nicolive.CommentOwner(cv.Lw.BroadID, http.MethodPut, &rq, cv.Ac)
				if err != nil {
					return err
				}
			} else {
				cv.Cmm.SendComment(ct.Text, ct.Iyayo)
			}

		case CommQueryAccountSet:
			if cv.Ac == nil {
				return nicolive.MakeError(nicolive.ErrOther, "Account data (cv.Ac) is nil.")
			}
			var ct CtQueryAccountSet
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}
			if ct.Mail != "" {
				cv.Ac.Mail = ct.Mail
			}
			if ct.Pass != "" {
				cv.Ac.Pass = ct.Pass
			}
			if ct.Usersession != "" {
				cv.Ac.Usersession = ct.Usersession
			}

			cv.AntennaConnect()

		case CommQueryAccountLogin:
			err := cv.Ac.Login()
			if err != nil {
				if nerr, ok := err.(nicolive.Error); ok {
					cv.CreateEvNewNotification(CtUINotificationTypeWarn, "login error", nerr.Description())
				} else {
					cv.CreateEvNewNotification(CtUINotificationTypeWarn, "login error", err.Error())
				}
				return err
			}
			cv.cli.log.Println("logged in")
			cv.CreateEvNewNotification(CtUINotificationTypeInfo, "login succeeded", "login succeeded")

		case CommQueryAccountLoad:
			var err error
			cv.Ac, err = nicolive.AccountLoad(filepath.Join(cv.cli.SavePath, accountFileName))
			return err

		case CommQueryAccountSave:
			return cv.Ac.Save(filepath.Join(cv.cli.SavePath, accountFileName))

		case CommQueryLogPrint:
			var ct CtQueryLogPrint
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			cv.cli.log.Printf("plug[%s] %s\n", cv.PluginName(m.prgno), ct.Text)

		case CommQuerySettingsSetCurrent:
			var ct CtQuerySettingsSetCurrent
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			cv.Settings = SettingsSlot(ct)
			for _, p := range cv.Pgns {
				p.SetState(!cv.Settings.PluginDisable[p.Name])
			}

		case CommQuerySettingsSetAll:
			var ct CtQuerySettingsSetAll
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			cv.cli.SettingsSlots = SettingsSlots(ct)

		case CommQueryPlugEnable:
			var ct CtQueryPlugEnable
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			pl, err := cv.Plugin(ct.No)
			if err != nil {
				return err
			}
			pl.SetState(ct.Enable)
			if cv.Settings.PluginDisable == nil {
				cv.Settings.PluginDisable = make(map[string]bool)
			}
			cv.Settings.PluginDisable[pl.Name] = !ct.Enable

		case CommQueryUserSet:
			var ct nicolive.User // CtQueryUserSet
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Storing the user info failed", "JSON parse error")
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			err := cv.prcdnle.userDB.Store(&ct)
			if err != nil {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Storing the user info failed", "DB error : "+err.Error())
				return err
			}

			cv.Evch <- NewMessageMust(DomainNagome, CommNagomeUserUpdate, CtNagomeUserUpdate(ct))

		case CommQueryUserSetName:
			var ct CtQueryUserSetName
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Storing the user name failed", "JSON parse error")
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			if ct.Name == "" {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Blank name", "You can't set blank name")
				return nicolive.MakeError(nicolive.ErrOther, "format error : Name is empty")
			}

			user, err := cv.prcdnle.userDB.Fetch(ct.ID)
			if err != nil {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Storing the user name failed", "DB error")
				return nicolive.MakeError(nicolive.ErrOther, "Storing the user name failed: DB error : "+err.Error())
			}
			if user == nil {
				user, err = cv.prcdnle.CheckIntervalAndCreateUser(ct.ID)
				if err != nil {
					cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Storing the user name failed", "DB error")
					return nicolive.MakeError(nicolive.ErrOther, "Storing the user name failed: DB error : "+err.Error())
				}
			}

			user.Name = ct.Name
			err = cv.prcdnle.userDB.Store(user)
			if err != nil {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Storing the user name failed", "DB error : "+err.Error())
				return err
			}

			cv.Evch <- NewMessageMust(DomainNagome, CommNagomeUserUpdate, CtNagomeUserUpdate(*user))

		case CommQueryUserDelete:
			var ct CtQueryUserDelete
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Deleting the user info failed", "JSON parse error")
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			err := cv.prcdnle.userDB.Remove(ct.ID)
			if err != nil {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Removing the user info failed", "DB error : "+err.Error())
				return err
			}

		case CommQueryUserFetch:
			var ct CtQueryUserFetch
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Fetching the user info failed", "JSON parse error")
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			user, err := cv.prcdnle.CheckIntervalAndCreateUser(ct.ID)
			if err != nil {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Removing the user info failed", "DB error : "+err.Error())
				return err
			}

			user_current, err := cv.prcdnle.userDB.Fetch(ct.ID)
			if err != nil {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Removing the user info failed", "DB error : "+err.Error())
				return err
			}
			if user_current == nil {
				// If the user was not in the DB
				user_current = user
			} else {
				user_current.Name = user.Name
				user_current.ThumbnailURL = user.ThumbnailURL
			}

			err = cv.prcdnle.userDB.Store(user_current)
			if err != nil {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Removing the user info failed", "DB error : "+err.Error())
				return err
			}

			cv.Evch <- NewMessageMust(DomainNagome, CommNagomeUserUpdate, CtNagomeUserUpdate(*user_current))

		case CommDirectUserGet:
			var ct CtDirectUserGet
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			user, err := cv.prcdnle.userDB.Fetch(ct.ID)
			if err != nil {
				cv.CreateEvNewNotification(CtUINotificationTypeWarn, "Removing the user info failed", "DB error : "+err.Error())
				cv.cli.log.Printf("Removing the user info failed.\n DB error : %s", err.Error())
				return err
			}

			cv.Evch <- NewMessageMust(DomainDirectngm, CommDirectngmUserGet, CtDirectngmUserGet(*user))

		default:
			return nicolive.MakeError(nicolive.ErrOther, "Message : invalid query command : "+m.Command)
		}

	case DomainAntenna:
		switch m.Command {
		case CommAntennaGot:
			if cv.Settings.AutoFollowNextWaku {
				var ct CtAntennaGot
				if err := json.Unmarshal(m.Content, &ct); err != nil {
					return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
				}
				if cv.Lw != nil && cv.Lw.Stream.CommunityID == ct.CommunityID {
					ct := CtQueryBroadConnect{ct.BroadID, 0}
					cv.cli.log.Println("following to " + ct.BroadID)
					cv.Evch <- NewMessageMust(DomainQuery, CommQueryBroadConnect, ct)
				}
			}
		}
	}

	return nil
}

func processDirectMessage(cv *CommentViewer, m *Message) error {
	switch m.Command {
	case CommDirectPlugList:
		c := CtDirectngmPlugList{&cv.Pgns}
		t, err := NewMessage(DomainDirectngm, CommDirectngmPlugList, c)
		if err != nil {
			return nicolive.ErrFromStdErr(err)
		}
		cv.Pgns[m.prgno].WriteMess(t)
	case CommDirectSettingsCurrent:
		t, err := NewMessage(DomainDirectngm, CommDirectngmSettingsCurrent, CtDirectngmSettingsCurrent(cv.Settings))
		if err != nil {
			return nicolive.ErrFromStdErr(err)
		}
		cv.Pgns[m.prgno].WriteMess(t)
	case CommDirectSettingsAll:
		t, err := NewMessage(DomainDirectngm, CommDirectngmSettingsAll, CtDirectngmSettingsAll(cv.cli.SettingsSlots))
		if err != nil {
			return nicolive.ErrFromStdErr(err)
		}
		cv.Pgns[m.prgno].WriteMess(t)
	default:
		return nicolive.MakeError(nicolive.ErrOther, "Message : invalid query command : "+m.Command)
	}

	return nil
}
