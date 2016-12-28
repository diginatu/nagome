package main

import (
	"context"
	"encoding/json"
	"log"
	"path/filepath"
	"regexp"

	"github.com/diginatu/nagome/nicolive"
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
				cv.CreateEvNewDialog(CtUIDialogTypeWarn, "invalid BroadID", "no valid BroadID found in the ID text")
				return nicolive.MakeError(nicolive.ErrOther, "no valid BroadID found in the ID text")
			}

			lw := &nicolive.LiveWaku{Account: cv.Ac, BroadID: broadMch}
			if err = lw.FetchInformation(); err != nil {
				return err
			}

			cv.Disconnect()

			cv.Lw = lw
			cv.Cmm, err = nicolive.CommentConnect(context.TODO(), *cv.Lw, cv.prcdnle)
			if err != nil {
				return err
			}
			log.Println("connecting")

		case CommQueryBroadDisconnect:
			cv.Disconnect()

		case CommQueryBroadSendComment:
			var ct CtQueryBroadSendComment
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}
			cv.Cmm.SendComment(ct.Text, ct.Iyayo)

		case CommQueryAccountSet:
			var ct CtQueryAccountSet
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}
			nicoac := nicolive.Account(ct)
			cv.Ac = &nicoac

			cv.AntennaConnect()

		case CommQueryAccountLogin:
			err := cv.Ac.Login()
			if err != nil {
				if nerr, ok := err.(nicolive.Error); ok {
					cv.CreateEvNewDialog(CtUIDialogTypeWarn, "login error", nerr.Description())
				} else {
					cv.CreateEvNewDialog(CtUIDialogTypeWarn, "login error", err.Error())
				}
				return err
			}
			log.Println("logged in")
			cv.CreateEvNewDialog(CtUIDialogTypeInfo, "login succeeded", "login succeeded")

		case CommQueryAccountLoad:
			return cv.Ac.Load(filepath.Join(App.SavePath, accountFileName))

		case CommQueryAccountSave:
			return cv.Ac.Save(filepath.Join(App.SavePath, accountFileName))

		case CommQueryLogPrint:
			var ct CtQueryLogPrint
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			log.Printf("plug[%s] %s\n", cv.Pgns[m.prgno].Name, ct.Text)

		case CommQuerySettingsSet:
			var ct CtQuerySettingsSet
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			cv.Settings = SettingsSlot(ct)

		case CommQuerySettingsSetAll:
			var ct CtQuerySettingsSetSlots
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			App.SettingsSlots = SettingsSlots(ct)

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
					ct := CtQueryBroadConnect{ct.BroadID}
					log.Println("following to " + ct.BroadID)
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
		t, err := NewMessage(DomainDirectngm, CommDirectngmSettingsAll, CtDirectngmSettingsAll(App.SettingsSlots))
		if err != nil {
			return nicolive.ErrFromStdErr(err)
		}
		cv.Pgns[m.prgno].WriteMess(t)
	default:
		return nicolive.MakeError(nicolive.ErrOther, "Message : invalid query command : "+m.Command)
	}

	return nil
}
