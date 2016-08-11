package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/diginatu/nagome/nicolive"
)

const (
	pluginFlashWaitDu time.Duration = 200 * time.Millisecond

	pluginNameMain string = "main"
)

type plugin struct {
	Name        string
	Description string
	Version     string
	Auther      string
	Exec        string
	Method      string
	Nagomever   string
	Depends     []string
	Rw          *bufio.ReadWriter
	FlushTm     *time.Timer
}

func (pl *plugin) depend(pln string) bool {
	f := false
	for _, d := range pl.Depends {
		if d == pln {
			f = true
			break
		}
	}
	return f
}

// eachPluginRw manages plugins IO. The number of its go routines is same as loaded plugins.
func eachPluginRw(cv *commentViewer, n int, wg *sync.WaitGroup) {
	defer wg.Done()

	dec := json.NewDecoder(cv.Pgns[n].Rw)
	mes := make(chan (*Message))

	// Run decoder.  It puts a message into "mes".
	go func() {
		for {
			m := new(Message)
			if err := dec.Decode(m); err != nil {
				if err != io.EOF {
					// TODO: emit error
					Logger.Println(err)
				}
				m = nil
			}

			select {
			case mes <- m:
			case <-cv.Quit:
				return
			}
		}
	}()

	for {
		select {
		// Process the message
		case m := <-mes:
			if m == nil {
				// quit if UI plugin disconnect
				if cv.Pgns[n].Name == pluginNameMain {
					close(cv.Quit)
				}
				return
			}

			Logger.Println("plugin message [", cv.Pgns[n].Name, "] : ", m)
			nicoerr := processPluginMessage(cv, m)
			if nicoerr != nil {
				Logger.Println("plugin message error [", cv.Pgns[n].Name, "] : ", nicoerr)
			}

		// Flush plugin IO
		case <-cv.Pgns[n].FlushTm.C:
			Logger.Println("plugin ", n, " flushing")
			cv.Pgns[n].Rw.Flush()

		case <-cv.Quit:
			return
		}
	}
}

func sendPluginEvent(cv *commentViewer, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {

		case mes := <-cv.Evch:
			jmes, err := json.Marshal(mes)
			if err != nil {
				Logger.Println(err)
				continue
			}
			for _, plug := range cv.Pgns {
				if plug.depend(mes.Domain) {
					_, err := fmt.Fprintf(plug.Rw.Writer, "%s\n", jmes)
					if err != nil {
						// TODO: emit error

						Logger.Println(err)
						continue
					}
					plug.FlushTm.Reset(pluginFlashWaitDu)
				}
			}
		case <-cv.Quit:
			return
		}

	}
}

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
