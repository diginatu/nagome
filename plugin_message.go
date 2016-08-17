package main

import (
	"encoding/json"
	"path/filepath"
	"regexp"

	"github.com/diginatu/nagome/nicolive"
)

func processPluginMessage(cv *commentViewer, m *Message) nicolive.NicoError {
	if m.Domain == "Nagome" {
		switch m.Func {

		case FuncQueryBroad:
			switch m.Command {

			case CommQueryBroadConnect:
				var ct CtQueryBroadConnect
				if err := json.Unmarshal(m.Content, &ct); err != nil {
					return nicolive.NicoErr(nicolive.NicoErrOther,
						"JSON error in the content", err.Error())
				}

				brdRg := regexp.MustCompile("(lv|co)\\d+")
				broadMch := brdRg.FindString(ct.BroadID)
				if broadMch == "" {
					cv.createEvNewDialog(CtUIDialogTypeWarn,
						"invalid BroadID", "no valid BroadID found in the ID text")
					return nicolive.NicoErr(nicolive.NicoErrOther,
						"invalid BroadID", "no valid BroadID found in the ID text")
				}

				cv.Lw = &nicolive.LiveWaku{Account: cv.Ac, BroadID: broadMch}

				if nicoerr := cv.Lw.FetchInformation(); nicoerr != nil {
					return nicoerr
				}
				if cv.Cmm.IsConnected {
					Logger.Println("Connected")
					if nicoerr := cv.Cmm.Disconnect(); nicoerr != nil {
						Logger.Println("discon err")
						return nicoerr
					}
					Logger.Println("disconnected")
				}
				if nicoerr := cv.Cmm.SetLv(cv.Lw); nicoerr != nil {
					return nicoerr
				}
				if nicoerr := cv.Cmm.Connect(); nicoerr != nil {
					return nicoerr
				}
				Logger.Println("connecting")

			case CommQueryBroadSendComment:
				var ct CtQueryBroadSendComment
				if err := json.Unmarshal(m.Content, &ct); err != nil {
					return nicolive.NicoErr(nicolive.NicoErrOther,
						"JSON error in the content", err.Error())
				}
				nicoerr := cv.Cmm.SendComment(ct.Text, ct.Iyayo)
				if nicoerr != nil {
					cv.createEvNewDialog(CtUIDialogTypeWarn, nicoerr.Code(), nicoerr.Description())
					return nicoerr
				}

			case CommQueryBroadDisconnect:
				cv.Cmm.Disconnect()

			default:
				return nicolive.NicoErr(nicolive.NicoErrOther,
					"Message", "invalid Command in received message")
			}

		case FuncQueryAccount:
			switch m.Command {
			case CommQueryAccountSet:
				var ct CtQueryAccountSet
				if err := json.Unmarshal(m.Content, &ct); err != nil {
					return nicolive.NicoErr(nicolive.NicoErrOther,
						"JSON error in the content", err.Error())
				}

				nicoac := nicolive.Account(ct)
				cv.Ac = &nicoac
			case CommQueryAccountLogin:
				nicoerr := cv.Ac.Login()
				if nicoerr != nil {
					cv.createEvNewDialog(CtUIDialogTypeWarn,
						"login error", nicoerr.Description())
					return nicoerr
				}
				Logger.Println("logged in")
				cv.createEvNewDialog(CtUIDialogTypeInfo,
					"login succeeded", "login succeeded")

			case CommQueryAccountSave:
				cv.Ac.Save(filepath.Join(App.SavePath, "userData.yml"))

			case CommQueryAccountLoad:
				cv.Ac.Load(filepath.Join(App.SavePath, "userData.yml"))

			default:
				return nicolive.NicoErr(nicolive.NicoErrOther,
					"Message", "invalid Command in received message")
			}

		case FuncComment:
			cv.Evch <- m

		default:
			return nicolive.NicoErr(nicolive.NicoErrOther,
				"Message", "invalid Func in received message")
		}
	} else {
		cv.Evch <- m
	}

	return nil
}
