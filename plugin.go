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

func (cv *commentViewer) readPluginMes(n int, wg *sync.WaitGroup) {
	defer wg.Done()
	decoded := make(chan bool)

	dec := json.NewDecoder(cv.Pgns[n].Rw)
	for {
		m := new(Message)
		go func() {
			if err := dec.Decode(m); err == io.EOF {
				decoded <- false
				return
			} else if err != nil {
				Logger.Println(err)
				decoded <- false
				return
			}
			decoded <- true
		}()

		select {
		case st := <-decoded:
			if !st {
				// quit if UI plugin disconnect
				if cv.Pgns[n].Name == "main" {
					close(cv.Quit)
				}
				return
			}
		case <-cv.Quit:
			return
		}

		if m.Domain == "Nagome" {
			switch m.Func {

			case FuncQueryBroad:
				switch m.Command {

				case CommQueryBroadConnect:
					var ct CtQueryBroadConnect
					if err := json.Unmarshal(m.Content, &ct); err != nil {
						Logger.Println("error in content:", err)
						continue
					}

					brdRg := regexp.MustCompile("(lv|co)\\d+")
					broadMch := brdRg.FindString(ct.BroadID)
					if broadMch == "" {
						Logger.Println("invalid BroadID")
						continue
					}

					cv.Lw = &nicolive.LiveWaku{Account: cv.Ac, BroadID: broadMch}

					nicoerr := cv.Lw.FetchInformation()
					if nicoerr != nil {
						Logger.Println(nicoerr)
						continue
					}

					eventReceiver := &commentEventEmit{cv: cv}
					cv.Cmm = nicolive.NewCommentConnection(cv.Lw, eventReceiver)
					nicoerr = cv.Cmm.Connect()
					if nicoerr != nil {
						Logger.Println(nicoerr)
						continue
					}
					defer cv.Cmm.Disconnect()

				case CommQueryBroadSendComment:
					var ct CtQueryBroadSendComment
					if err := json.Unmarshal(m.Content, &ct); err != nil {
						Logger.Println("error in content:", err)
						continue
					}
					cv.Cmm.SendComment(ct.Text, ct.Iyayo)

				default:
					Logger.Println("invalid Command in received message")
				}

			case FuncQueryAccount:
				switch m.Command {
				case CommQueryAccountSet:
					var ct CtQueryAccountSet
					if err := json.Unmarshal(m.Content, &ct); err != nil {
						Logger.Println("error in content:", err)
						continue
					}

					nicoac := nicolive.Account(ct)
					cv.Ac = &nicoac
				case CommQueryAccountLogin:
					err := cv.Ac.Login()
					if err != nil {
						Logger.Println(err)
						t, err := NewMessage(DomainNagome, FuncUI, CommUIDialog,
							CtUIDialog{
								Type:        "warn",
								Title:       "login error",
								Description: err.Description(),
							})
						Logger.Println(t)
						if err != nil {
							Logger.Println(err)
							continue
						}
						cv.Evch <- t
						continue
					}
					Logger.Println("logged in")

				case CommQueryAccountSave:
					cv.Ac.Save(filepath.Join(App.SavePath, "userData.yml"))

				case CommQueryAccountLoad:
					cv.Ac.Load(filepath.Join(App.SavePath, "userData.yml"))

				default:
					Logger.Println("invalid Command in received message")
				}

			case FuncComment:
				cv.Evch <- m

			default:
				Logger.Println("invalid Func in received message")
			}
		} else {
			cv.Evch <- m
		}
	}
}

func (cv *commentViewer) sendPluginEvent(i int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case mes := <-cv.Evch:
			jmes, _ := json.Marshal(mes)
			for _, plug := range cv.Pgns {
				if plug.depend(mes.Domain) {
					_, err := fmt.Fprintf(plug.Rw.Writer, "%d %s\n", i, jmes)
					if err != nil {
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

func (cv *commentViewer) flushPluginIO(i int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-cv.Pgns[i].FlushTm.C:
			Logger.Println("plugin ", i, " flushing")
			cv.Pgns[i].Rw.Flush()
		case <-cv.Quit:
			return
		}
	}
}
