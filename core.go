package main

import (
	"bufio"
	"encoding/json"
	"io"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/diginatu/nagome/nicolive"
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
}

// A commentViewer is a pair of an Account and a LiveWaku.
type commentViewer struct {
	Ac   *nicolive.Account
	Lw   *nicolive.LiveWaku
	Cmm  *nicolive.CommentConnection
	Pgns []*plugin
}

func (cv *commentViewer) runCommentViewer() error {
	var wg sync.WaitGroup

	wg.Add(len(cv.Pgns))
	for i, pg := range cv.Pgns {
		Logger.Println(pg.Name)
		go readPluginMes(cv, i, &wg)
	}

	wg.Wait()
	return nil
}

func readPluginMes(cv *commentViewer, n int, wg *sync.WaitGroup) {
	defer wg.Done()

	dec := json.NewDecoder(cv.Pgns[n].Rw)
	for {
		var m Message
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			Logger.Println(err)
			return
		}

		if m.Domain == "Nagome" {
			switch m.Func {

			case FuncnBroadQuery:
				switch m.Command {

				case CommBroadQueryConnect:
					var ct CtBroadQueryConnect
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

					cv.Cmm = nicolive.NewCommentConnection(cv.Lw, nil)
					nicoerr = cv.Cmm.Connect()
					if nicoerr != nil {
						Logger.Println(nicoerr)
						continue
					}

					defer cv.Cmm.Disconnect()

				case CommBroadQuerySendComment:
					var ct CtBroadQuerySendComment
					if err := json.Unmarshal(m.Content, &ct); err != nil {
						Logger.Println("error in content:", err)
						continue
					}
					cv.Cmm.SendComment(ct.Text, ct.Iyayo)

				default:
					Logger.Println("invalid Command in received message")
				}

			case FuncnAccountQuery:
				switch m.Command {
				case CommAccountLogin:
					err := cv.Ac.Login()
					if err != nil {
						Logger.Fatalln(err)
						continue
					}
					Logger.Println("logged in")

				case CommAccountSave:
					cv.Ac.Save(filepath.Join(App.SavePath, "userData.yml"))

				case CommAccountLoad:
					cv.Ac.Load(filepath.Join(App.SavePath, "userData.yml"))

				default:
					Logger.Println("invalid Command in received message")
				}

			default:
				Logger.Println("invalid Func in received message")
			}
		}
	}

	return
}
