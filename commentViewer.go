package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/diginatu/nagome/nicolive"
)

// commentEventEmit will receive comment events and emits commentViewer events.
type commentEventEmit struct {
	cv *commentViewer
}

func (der *commentEventEmit) Proceed(ev *nicolive.Event) {
	var content []byte
	var command string

	switch ev.Type {
	case nicolive.EventTypeGot:
		content, _ = json.Marshal(ev.Content.(nicolive.Comment))
		command = CommCommentAdd
	default:
		log.Println(ev.String())
	}

	if command != "" {
		der.cv.Evch <- &Message{
			Domain:  DomainNagome,
			Func:    FuncComment,
			Command: command,
			Content: content,
		}
	}
}

// A commentViewer is a pair of an Account and a LiveWaku.
type commentViewer struct {
	Ac      *nicolive.Account
	Lw      *nicolive.LiveWaku
	Cmm     *nicolive.CommentConnection
	Pgns    []*plugin
	TCPPort string
	Evch    chan *Message
	Quit    chan struct{}
}

func (cv *commentViewer) Run() {
	defer cv.Cmm.Disconnect()

	cv.loadPlugins()

	var wg sync.WaitGroup

	wg.Add(1)
	go pluginTCPServer(cv, &wg)

	wg.Add(1)
	go sendPluginEvent(cv, &wg)

	wg.Add(len(cv.Pgns))
	for i, pg := range cv.Pgns {
		log.Println(pg.Name)
		go eachPluginRw(cv, i, &wg)
	}

	wg.Wait()

	return
}

func (cv *commentViewer) AddPlugin(p *plugin) {
	cv.Pgns = append(cv.Pgns, p)
	p.Init(len(cv.Pgns))
}

func (cv *commentViewer) loadPlugins() error {
	psPath := filepath.Join(App.SavePath, pluginDirName)

	ds, err := ioutil.ReadDir(psPath)
	if err != nil {
		return err
	}

	for _, d := range ds {
		if d.IsDir() {
			fmt.Println(d.Name())
			p := new(plugin)
			pPath := filepath.Join(psPath, d.Name())
			err = p.loadPlugin(filepath.Join(pPath, "plugin.yml"))
			if err != nil {
				log.Println("failed load plugin : ", d.Name())
				log.Println(err)
				continue
			}

			cv.AddPlugin(p)

			for i := range p.Exec {
				p.Exec[i] = strings.Replace(p.Exec[i], "{{path}}", pPath, -1)
				p.Exec[i] = strings.Replace(p.Exec[i], "{{port}}", cv.TCPPort, -1)
				p.Exec[i] = strings.Replace(p.Exec[i], "{{no}}", strconv.Itoa(p.No()), -1)
			}

			switch p.Method {
			case pluginMethodTCP:
				if len(p.Exec) > 1 {
					cmd := exec.Command(p.Exec[0], p.Exec[1:]...)
					err := cmd.Start()
					if err != nil {
						log.Fatal(err)
					}
				}
			case pluginMethodStd:
			default:
			}

			log.Println(p)
		}
	}

	return nil
}

func (cv *commentViewer) CreateEvNewDialog(typ, title, desc string) {
	t, err := NewMessage(DomainNagome, FuncUI, CommUIDialog,
		CtUIDialog{
			Type:        typ,
			Title:       title,
			Description: desc,
		})
	if err != nil {
		log.Println(err)
	}
	cv.Evch <- t
}
