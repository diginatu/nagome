package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	pluginFlashWaitDu time.Duration = 200 * time.Millisecond

	pluginNameMain  string = "main"
	pluginMethodTCP        = "tcp"
	pluginMethodStd        = "std"
)

type plugin struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Version     string            `yaml:"version"`
	Author      string            `yaml:"author"`
	Method      string            `yaml:"method"`
	Exec        []string          `yaml:"exec"`
	Nagomever   string            `yaml:"nagomever"`
	Depends     []string          `yaml:"depends"`
	Rw          *bufio.ReadWriter `yaml:"-"`
	Enablc      chan struct{}     `yaml:"-"`
	flushTm     *time.Timer
	no          int
}

func (pl *plugin) Init(no int) {
	pl.flushTm = time.NewTimer(time.Hour)
	pl.Enablc = make(chan struct{}, 1)
	pl.no = no
}

func (pl *plugin) Enable(cv *CommentViewer) {
	if pl.no == 0 {
		log.Printf("plugin \"%s\" is not initialized\n", pl.Name)
		return
	}
	if pl.Name == "" {
		log.Printf("plugin \"%s\" no name is set\n", pl.Name)
		return
	}
	pl.Enablc <- struct{}{}

	cv.wg.Add(1)
	go eachPluginRw(cv, pl.no-1)

	return
}

func (pl *plugin) Depend(pln string) bool {
	f := false
	for _, d := range pl.Depends {
		if d == pln {
			f = true
			break
		}
	}
	return f
}

func (pl *plugin) No() int {
	return pl.no
}

func (pl *plugin) loadPlugin(filePath string) error {
	d, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(d, pl)
	if err != nil {
		return err
	}

	return nil
}

func (pl *plugin) savePlugin(filePath string) error {
	d, err := yaml.Marshal(pl)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, d, 0600)
	if err != nil {
		return err
	}

	return nil
}

// eachPluginRw manages plugins IO. It is launched when a plugin is leaded.
func eachPluginRw(cv *CommentViewer, n int) {
	defer cv.wg.Done()

	// wait for being enabled
	select {
	case <-cv.Pgns[n].Enablc:
	case <-cv.Quit:
		return
	}

	// Run decoder.  It puts a message into "mes".
	dec := json.NewDecoder(cv.Pgns[n].Rw)
	mes := make(chan (*Message))
	go func() {
		for {
			m := new(Message)
			if err := dec.Decode(m); err != nil {
				if err != io.EOF {
					select {
					// ignore if quitting
					case <-cv.Quit:
					default:
						cv.CreateEvNewDialog(CtUIDialogTypeInfo, "plugin discconect",
							fmt.Sprintf("plugin [%s] : connection desconnected", cv.Pgns[n].Name))
						log.Println(err)
					}
				}
				cv.Pgns[n].Rw = nil
				m = nil
			}

			select {
			case mes <- m:
				if m == nil {
					return
				}
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

			log.Println("plugin message [", cv.Pgns[n].Name, "] : ", m)
			nicoerr := processPluginMessage(cv, m)
			if nicoerr != nil {
				log.Println("plugin message error [", cv.Pgns[n].Name, "] : ", nicoerr)
			}

		// Flush plugin IO
		case <-cv.Pgns[n].flushTm.C:
			log.Println("plugin ", n, " flushing")
			cv.Pgns[n].Rw.Flush()

		case <-cv.Quit:
			return
		}
	}
}

func sendPluginEvent(cv *CommentViewer) {
	defer cv.wg.Done()

	for {
		select {

		case mes := <-cv.Evch:
			jmes, err := json.Marshal(mes)
			if err != nil {
				log.Println(err)
				continue
			}
			for _, plug := range cv.Pgns {
				if plug.Rw != nil && plug.Depend(mes.Domain) {
					_, err := fmt.Fprintf(plug.Rw.Writer, "%s\n", jmes)
					if err != nil {
						// TODO: emit error

						log.Println(err)
						continue
					}
					plug.flushTm.Reset(pluginFlashWaitDu)
				}
			}
		case <-cv.Quit:
			return
		}

	}
}

func pluginTCPServer(cv *CommentViewer) {
	defer cv.wg.Done()

	adr, err := net.ResolveTCPAddr("tcp", ":"+cv.TCPPort)
	if err != nil {
		log.Panicln(err)
	}
	l, err := net.ListenTCP("tcp", adr)
	if err != nil {
		log.Panicln(err)
	}
	defer l.Close()

	for {
		l.SetDeadline(time.Now().Add(time.Second))
		select {
		default:
			conn, err := l.Accept()
			if err != nil {
				nerr, ok := err.(net.Error)
				if ok && nerr.Timeout() && nerr.Temporary() {
					continue
				}
				log.Println(err)
				continue
			}
			cv.wg.Add(1)
			go handleTCPPlugin(conn, cv)
		case <-cv.Quit:
			return
		}
	}
}

func handleTCPPlugin(c net.Conn, cv *CommentViewer) {
	defer cv.wg.Done()
	defer c.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))

	errc := make(chan struct{})

	cv.wg.Add(1)
	go func() {
		defer cv.wg.Done()
		for {
			select {
			default:
				dec := json.NewDecoder(rw)
				var ct CtQueryPluginNo
				err := dec.Decode(&ct)
				if err != nil {
					// ignore if quitting
					select {
					case <-cv.Quit:
					default:
						log.Println(err)
					}
					close(errc)
					return
				}

				n := ct.No - 1
				if n < 0 || n >= len(cv.Pgns) {
					log.Println("received invalid plugin No.")
					close(errc)
					return
				}
				if cv.Pgns[n].Rw != nil {
					log.Println("plugin is already connected")
					close(errc)
					return
				}
				cv.Pgns[n].Rw = rw
				cv.Pgns[n].Enable(cv)
				log.Println("loaded plugin ", cv.Pgns[n])
				break

			case <-cv.Quit:
				return
			}
			break
		}

	}()

	// wait for quitting or error in above go routine
	select {
	case <-errc:
	case <-cv.Quit:
	}
}

func handleSTDPlugin(p *plugin, cv *CommentViewer) {
	defer cv.wg.Done()

	if len(p.Exec) < 1 {
		log.Printf("exec is not specified in plugin [%s]\n", p.Name)
		return
	}

	cmd := exec.Command(p.Exec[0], p.Exec[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Println(err)
		return
	}
	defer stdin.Close()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return
	}
	defer stdout.Close()
	err = cmd.Start()
	if err != nil {
		log.Println(err)
		return
	}

	p.Rw = bufio.NewReadWriter(bufio.NewReader(stdout), bufio.NewWriter(stdin))
	p.Enable(cv)
	log.Println("loaded plugin ", p)

	<-cv.Quit
}
