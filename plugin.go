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
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	pluginFlashWaitDu time.Duration = 50 * time.Millisecond

	pluginMethodTCP string = "tcp"
	pluginMethodStd        = "std"
)

type plugin struct {
	Name        string            `yaml:"name"        json:"name"`
	Description string            `yaml:"description" json:"description"`
	Version     string            `yaml:"version"     json:"version"`
	Author      string            `yaml:"author"      json:"author"`
	Method      string            `yaml:"method"      json:"method"`
	Exec        []string          `yaml:"exec"        json:"-"`
	Nagomever   string            `yaml:"nagomever"   json:"-"`
	Depends     []string          `yaml:"depends"     json:"depends"`
	No          int               `yaml:"-"           json:"no"`
	Rw          *bufio.ReadWriter `yaml:"-"           json:"-"`
	Startc      chan struct{}     `yaml:"-"           json:"-"`
	flushTm     *time.Timer
	isEnable    bool
}

func (pl *plugin) Init(no int) {
	pl.flushTm = time.NewTimer(time.Hour)
	pl.Startc = make(chan struct{}, 1)
	pl.No = no
}

func (pl *plugin) Start(cv *CommentViewer) {
	if pl.No == 0 {
		log.Printf("plugin \"%s\" is not initialized\n", pl.Name)
		return
	}
	if pl.Name == "" {
		log.Printf("plugin \"%s\" no name is set\n", pl.Name)
		return
	}
	if pl.Rw == nil {
		log.Printf("plugin \"%s\" no rw\n", pl.Name)
		return
	}
	if pl.isEnable {
		return
	}
	pl.Enable()

	pl.Startc <- struct{}{}

	cv.wg.Add(1)
	go eachPluginRw(cv, pl.No-1)
}

func (pl *plugin) Enable() {
	if pl.isEnable {
		return
	}
	pl.isEnable = true

	// send message
	jmes, err := json.Marshal(Message{
		Domain:  DomainDirectngm,
		Command: CommDirectngmPlugEnabled,
	})
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Fprintf(pl.Rw, "%s\n", jmes)
	pl.flushTm.Reset(0)
}

func (pl *plugin) Disable() {
	if !pl.isEnable {
		return
	}
	pl.isEnable = false

	// send message
	jmes, err := json.Marshal(Message{
		Domain:  DomainDirectngm,
		Command: CommDirectngmPlugDisabled,
	})
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Fprintf(pl.Rw, "%s\n", jmes)
	pl.flushTm.Reset(0)
}

func (pl *plugin) IsEnable() bool {
	return pl.isEnable
}

func (pl *plugin) DependFilter(pln string) bool {
	f := false
	for _, d := range pl.Depends {
		if d == pln+DomainFilterSuffix {
			f = true
			break
		}
	}
	return f
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

func (pl *plugin) isMain() bool {
	return pl.No == 0
}

// eachPluginRw manages plugins IO. It is launched when a plugin is leaded.
func eachPluginRw(cv *CommentViewer, n int) {
	defer cv.wg.Done()

	// wait for being enabled
	select {
	case <-cv.Pgns[n].Startc:
	case <-cv.Quit:
		return
	}

	// Run decoder.  It puts a message into "mes".
	dec := json.NewDecoder(cv.Pgns[n].Rw)
	mes := make(chan (*Message))
	cv.wg.Add(1)
	go func() {
		defer cv.wg.Done()
		for {
			m := new(Message)
			err := dec.Decode(m)
			if err != nil {
				if err != io.EOF {
					select {
					// ignore if quitting
					case <-cv.Quit:
					default:
						cv.CreateEvNewDialog(CtUIDialogTypeInfo, "plugin disconnect",
							fmt.Sprintf("plugin [%s] : connection disconnected", cv.Pgns[n].Name))
						log.Println(err)
					}
				}
				cv.Pgns[n].Rw = nil
				m = nil
			} else {
				m.prgno = n
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
		if !cv.Pgns[n].IsEnable() {
			// wait for being enabled
			select {
			case <-cv.Pgns[n].Startc:
			case <-cv.Quit:
				return
			}
		}

		select {
		// Process received message
		case m := <-mes:
			if m == nil {
				// quit if UI plugin disconnect
				if cv.Pgns[n].isMain() {
					cv.Cmm.Disconnect()
					close(cv.Quit)
				}
				continue
			}

			// ignore if plugin is not enabled
			if cv.Pgns[n].IsEnable() {
				log.Printf("plugin message [%s] : %v", cv.Pgns[n].Name, m)
				cv.Evch <- m
			}

		// Flush plugin IO
		case <-cv.Pgns[n].flushTm.C:
			cv.Pgns[n].Rw.Flush()

		case <-cv.Quit:
			cv.Pgns[n].Rw = nil
			return
		}
	}
}

func sendPluginMessage(cv *CommentViewer) {
	defer cv.wg.Done()

	for {
	readLoop:
		select {
		case mes := <-cv.Evch:
			// Direct
			if mes.Domain == DomainDirectngm {
				plug := cv.Pgns[mes.prgno]
				if plug.Rw == nil {
					continue
				}
				jmes, err := json.Marshal(mes)
				if err != nil {
					log.Println(err)
					log.Println(mes)
					continue
				}
				plug.flushTm.Reset(pluginFlashWaitDu)
				_, err = fmt.Fprintf(plug.Rw, "%s\n", jmes)
				if err != nil {
					cv.CreateEvNewDialog(CtUIDialogTypeInfo, "plugin", "failed to send event : "+plug.Name)
					log.Println(err)

					plug.Rw = nil
					continue
				}
				continue
			}
			if mes.Domain == DomainDirect {
				go func() {
					nicoerr := processDirectMessage(cv, mes)
					if nicoerr != nil {
						log.Printf("plugin message error form [%s] : %s\n", cv.Pgns[mes.prgno].Name, nicoerr)
						log.Println(mes)
					}
				}()
				continue
			}

			// filter

			// Messages from filter plugin will not send same plugin.
			var st int
			if strings.HasSuffix(mes.Domain, DomainFilterSuffix) {
				st = mes.prgno + 1
				mes.Domain = strings.TrimSuffix(mes.Domain, DomainFilterSuffix)
			}
			for i := st; i < len(cv.Pgns); i++ {
				plug := cv.Pgns[i]

				if plug.Rw != nil && plug.DependFilter(mes.Domain) {
					plug.flushTm.Reset(pluginFlashWaitDu)

					// A message to filter plugin has filter domain.
					tmes := *mes
					tmes.Domain = DomainFilterSuffix + mes.Domain
					jmes, err := json.Marshal(tmes)
					if err != nil {
						log.Println(err)
						log.Println(mes)
						break readLoop
					}
					plug.flushTm.Reset(pluginFlashWaitDu)
					_, err = fmt.Fprintf(plug.Rw, "%s\n", jmes)
					if err != nil {
						cv.CreateEvNewDialog(CtUIDialogTypeInfo, "plugin", "failed to send event : "+plug.Name)
						log.Println(err)

						plug.Rw = nil
						continue
					}
					break readLoop
				}
			}

			jmes, err := json.Marshal(mes)
			if err != nil {
				log.Println(err)
				log.Println(mes)
				continue
			}

			var wg sync.WaitGroup

			// regular
			for i := range cv.Pgns {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					plug := cv.Pgns[i]

					if plug.Rw != nil && plug.Depend(mes.Domain) {
						plug.flushTm.Reset(pluginFlashWaitDu)
						_, err := fmt.Fprintf(plug.Rw, "%s\n", jmes)
						if err != nil {
							cv.CreateEvNewDialog(CtUIDialogTypeInfo, "plugin", "failed to send event : "+plug.Name)
							log.Println(err)

							plug.Rw = nil
							return
						}
					}
				}(i)
			}

			go func() {
				nicoerr := processPluginMessage(cv, mes)
				if nicoerr != nil {
					log.Printf("plugin message error form [%s] : %s\n", cv.Pgns[mes.prgno].Name, nicoerr)
					log.Println(mes)
				}
			}()

			wg.Wait()

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
				errf := func(s interface{}) {
					// ignore if quitting
					select {
					case <-cv.Quit:
					default:
						log.Println(s)
					}
					close(errc)
				}

				dec := json.NewDecoder(rw)
				m := new(Message)
				err := dec.Decode(m)
				if err != nil {
					errf(err)
					return
				}
				if m.Domain != DomainDirect || m.Command != CommDirectNo {
					errf("send Direct.No message at first")
					return
				}

				var ct CtDirectNo
				if err := json.Unmarshal(m.Content, &ct); err != nil {
					errf(err)
					return
				}

				n := ct.No - 1
				if n < 0 || n >= len(cv.Pgns) {
					errf("received invalid plugin No.")
					return
				}
				if cv.Pgns[n].Rw != nil {
					errf("plugin is already connected")
					return
				}
				cv.Pgns[n].Rw = rw
				cv.Pgns[n].Start(cv)
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
	p.Start(cv)
	log.Println("loaded plugin ", p)

	<-cv.Quit
}
