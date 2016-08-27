package main

import (
	"encoding/json"
	"log"
	"path/filepath"
	"regexp"

	"github.com/diginatu/nagome/nicolive"
)

func processPluginMessage(cv *CommentViewer, m *Message) nicolive.Error {
	if m.Domain != DomainQuery {
		return nil
	}

	switch m.Command {
	case CommQueryBroadConnect:
		var ct CtQueryBroadConnect
		if err := json.Unmarshal(m.Content, &ct); err != nil {
			return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
		}

		brdRg := regexp.MustCompile("(lv|co)\\d+")
		broadMch := brdRg.FindString(ct.BroadID)
		if broadMch == "" {
			cv.CreateEvNewDialog(CtUIDialogTypeWarn, "invalid BroadID", "no valid BroadID found in the ID text")
			return nicolive.MakeError(nicolive.ErrOther, "no valid BroadID found in the ID text")
		}

		cv.Lw = &nicolive.LiveWaku{Account: cv.Ac, BroadID: broadMch}

		if nicoerr := cv.Lw.FetchInformation(); nicoerr != nil {
			return nicoerr
		}
		if cv.Cmm.IsConnected {
			log.Println("Connected")
			if nicoerr := cv.Cmm.Disconnect(); nicoerr != nil {
				log.Println("discon err")
				return nicoerr
			}
			log.Println("disconnected")
		}
		if nicoerr := cv.Cmm.SetLv(cv.Lw); nicoerr != nil {
			return nicoerr
		}
		if nicoerr := cv.Cmm.Connect(); nicoerr != nil {
			return nicoerr
		}
		log.Println("connecting")

	case CommQueryBroadDisconnect:
		cv.Cmm.Disconnect()

	case CommQueryBroadSendComment:
		var ct CtQueryBroadSendComment
		if err := json.Unmarshal(m.Content, &ct); err != nil {
			return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
		}
		nicoerr := cv.Cmm.SendComment(ct.Text, ct.Iyayo)
		if nicoerr != nil {
			// TODO: make dialog from nicolive.Error
			//cv.CreateEvNewDialog(CtUIDialogTypeWarn, nicoerr.Code(), nicoerr.Description())
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

	default:
		return nicolive.MakeError(nicolive.ErrOther, "Message : invalid query command")
	}

	return nil
}
