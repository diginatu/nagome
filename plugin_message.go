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
	broadIDRegex = regexp.MustCompile("(lv|co)\\d+")
)

func processPluginMessage(cv *CommentViewer, m *Message) error {
	if m.Domain != DomainQuery {
		return nil
	}

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

		cv.Lw = &nicolive.LiveWaku{Account: cv.Ac, BroadID: broadMch}

		if err = cv.Lw.FetchInformation(); err != nil {
			return err
		}

		cv.Disconnect()

		cv.Cmm, err = nicolive.CommentConnect(context.TODO(), cv.Lw, NewProceedNicoliveEvent(cv))
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
		nicoerr := cv.Cmm.SendComment(ct.Text, ct.Iyayo)
		if nicoerr != nil {
			cv.CreateEvNewDialog(CtUIDialogTypeWarn, "Send comment error", nicoerr.Description())
			return nicoerr
		}

	case CommQueryAccountSet:
		var ct CtQueryAccountSet
		if err := json.Unmarshal(m.Content, &ct); err != nil {
			return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
		}

		nicoac := nicolive.Account(ct)
		cv.Ac = &nicoac

	case CommQueryAccountLogin:
		nicoerr := cv.Ac.Login()
		if nicoerr != nil {
			cv.CreateEvNewDialog(CtUIDialogTypeWarn,
				"login error", nicoerr.Description())
			return nicoerr
		}
		log.Println("logged in")
		cv.CreateEvNewDialog(CtUIDialogTypeInfo,
			"login succeeded", "login succeeded")

	case CommQueryAccountLoad:
		cv.Ac.Load(filepath.Join(App.SavePath, accountFileName))

	case CommQueryAccountSave:
		cv.Ac.Save(filepath.Join(App.SavePath, accountFileName))

	case CommQueryLogPrint:
		var ct CtQueryLogPrint
		if err := json.Unmarshal(m.Content, &ct); err != nil {
			return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
		}

		log.Printf("plug[%s] %s\n", cv.Pgns[m.prgno-1].Name, ct.Text)

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

		log.Printf("%s\n", m.Content)
		log.Printf("%v\n", *ct.Config[0])

		App.SettingsSlots = SettingsSlots(ct)

	default:
		return nicolive.MakeError(nicolive.ErrOther, "Message : invalid query command : "+m.Command)
	}

	return nil
}

func processDirectMessage(cv *CommentViewer, m *Message) nicolive.Error {
	switch m.Command {
	case CommDirectPlugList:
		c := CtDirectngmPlugList{&cv.Pgns}
		t, err := NewMessage(DomainDirectngm, CommDirectngmPlugList, c)
		if err != nil {
			return nicolive.ErrFromStdErr(err)
		}
		t.prgno = m.prgno
		cv.Evch <- t
	case CommDirectSettingsCurrent:
		t, err := NewMessage(DomainDirectngm, CommDirectngmSettingsCurrent, cv.Settings)
		if err != nil {
			return nicolive.ErrFromStdErr(err)
		}
		t.prgno = m.prgno
		cv.Evch <- t
	case CommDirectSettingsAll:
		t, err := NewMessage(DomainDirectngm, CommDirectngmSettingsAll, App.SettingsSlots)
		if err != nil {
			return nicolive.ErrFromStdErr(err)
		}
		t.prgno = m.prgno
		cv.Evch <- t
	default:
		return nicolive.MakeError(nicolive.ErrOther, "Message : invalid query command : "+m.Command)
	}

	return nil
}
