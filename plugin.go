package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	pluginFlashWaitDu time.Duration = 50 * time.Millisecond

	pluginMethodTCP           = "tcp"
	pluginMethodStd           = "std"
	pluginEachMessageChanSize = 3
)

type pluginState int

const (
	pluginStateClose pluginState = iota
	pluginStateEnable
	pluginStateDisable
)

type plugin struct {
	Name        string      `yaml:"name"        json:"name"`
	Description string      `yaml:"description" json:"description"`
	Version     string      `yaml:"version"     json:"version"`
	Author      string      `yaml:"author"      json:"author"`
	Method      string      `yaml:"method"      json:"method"`
	Exec        []string    `yaml:"exec"        json:"-"`
	Nagomever   string      `yaml:"nagomever"   json:"-"`
	Subscribe   []string    `yaml:"subscribe"   json:"subscribe"`
	No          int         `yaml:"-"           json:"no"`
	GetState    pluginState `yaml:"-"           json:"state"` // Don't change directly
	setStateCh  chan (pluginState)
	stateMu     sync.Mutex
	rwc         io.ReadWriteCloser
	flushTm     *time.Timer
	wg          sync.WaitGroup
	cv          *CommentViewer
	quit        chan (struct{})
	writec      chan ([]byte)
}

// NewPlugin makes new Plugin.
func newPlugin(cv *CommentViewer) *plugin {
	return &plugin{
		No:         -1,
		quit:       make(chan struct{}),
		setStateCh: make(chan pluginState),
		writec:     make(chan []byte, pluginEachMessageChanSize),
		cv:         cv,
	}
}

func (pl *plugin) Open(rwc io.ReadWriteCloser, enable bool) error {
	pl.stateMu.Lock()
	defer pl.stateMu.Unlock()

	if pl.No == -1 {
		return fmt.Errorf("plugin \"%s\" is not initialized (add to CommentViewer)\n", pl.Name)
	}
	if pl.Name == "" {
		return fmt.Errorf("plugin \"%s\" no name is set\n", pl.Name)
	}
	if rwc == nil {
		return fmt.Errorf("given rw is nil\n")
	}
	if pl.GetState != pluginStateClose {
		return fmt.Errorf("already opened\n")
	}

	pl.rwc = rwc
	pl.flushTm = time.NewTimer(time.Minute)

	pl.wg.Add(1)
	go pl.evRoutine()

	var st pluginState
	if enable {
		st = pluginStateEnable
	} else {
		st = pluginStateDisable
	}
	pl.stateMu.Unlock()
	pl.setStateCh <- st
	pl.setStateCh <- st // wait for completing previous task
	pl.stateMu.Lock()

	return nil
}

func (pl *plugin) SetState(enable bool) {
	if pl.GetState == pluginStateClose {
		return
	}

	var st pluginState
	if enable {
		st = pluginStateEnable
	} else {
		st = pluginStateDisable
	}
	select {
	case pl.setStateCh <- st:
	case <-pl.quit:
		return
	}
}

func (pl *plugin) WriteMess(m *Message) (fail bool) {
	jm, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		log.Println(m)
		return
	}
	return pl.Write(jm)
}

func (pl *plugin) Write(p []byte) (fail bool) {
	pl.stateMu.Lock()
	defer pl.stateMu.Unlock()
	if pl.GetState != pluginStateEnable {
		return true
	}
	select {
	case pl.writec <- p:
		return false
	case <-pl.quit:
	}
	return true
}

func (pl *plugin) IsSubscribe(pln string) bool {
	f := false
	for _, d := range pl.Subscribe {
		if d == pln {
			f = true
			break
		}
	}
	return f
}

func (pl *plugin) Load(filePath string) error {
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

func (pl *plugin) Save(filePath string) error {
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

func (pl *plugin) IsMain() bool {
	return pl.No == 0
}

func (pl *plugin) evRoutine() {
	defer pl.wg.Done()
	defer func() {
		err := pl.rwc.Close()
		if err != nil {
			log.Println(err)
		}
		pl.stateMu.Lock()
		pl.GetState = pluginStateClose
		pl.stateMu.Unlock()
	}()
	defer log.Printf("plugin [%s] is closing", pl.Name)

	// Run decoder.  It puts a message into "mes".
	dec := json.NewDecoder(pl.rwc)
	mes := make(chan (*Message))
	pl.wg.Add(1)
	go func() {
		defer pl.wg.Done()
		for {
			m := new(Message)
			err := dec.Decode(m)
			if err != nil {
				select {
				// ignore if quitting
				case <-pl.quit:
				default:
					if err != io.EOF {
						pl.cv.CreateEvNewDialog(CtUIDialogTypeInfo, "plugin disconnected",
							fmt.Sprintf("plugin [%s] : connection disconnected", pl.Name))
						log.Println(err)
					}
				}
				m = nil
			}

			select {
			case mes <- m:
				if m == nil {
					return
				}
			case <-pl.quit:
				return
			}
		}
	}()

	bufw := bufio.NewWriter(pl.rwc)
	writeMess := func(p []byte) {
		pl.flushTm.Reset(pluginFlashWaitDu)
		_, err := fmt.Fprintf(bufw, "%s\n", p)
		if err != nil {
			log.Println(err)
			pl.cv.CreateEvNewDialog(CtUIDialogTypeInfo, "plugin", "failed to write a message : "+pl.Name)
			// quit if UI plugin disconnect
			if pl.IsMain() {
				pl.cv.Quit()
			} else {
				pl.close()
			}
		}
		return
	}
	for {
		select {
		// Process a received message
		case m := <-mes:
			if m == nil {
				// quit if main plugin is disconnected
				if pl.IsMain() {
					pl.cv.Quit()
				} else {
					pl.close()
				}
				continue
			}
			if pl.GetState != pluginStateEnable {
				continue
			}
			m.prgno = pl.No
			log.Printf("plugin message [%s] : %v", pl.Name, m)
			pl.cv.Evch <- m

		// Send a message
		case m := <-pl.writec:
			if pl.GetState != pluginStateEnable {
				continue
			}
			writeMess(m)

		// Flush plugin IO
		case <-pl.flushTm.C:
			err := bufw.Flush()
			if err != nil {
				log.Println(err)
				continue
			}

		case e := <-pl.setStateCh:
			func() {
				pl.stateMu.Lock()
				defer pl.stateMu.Unlock()

				if e == pluginStateClose {
					return
				}
				if pl.GetState == e {
					return
				}
				pl.GetState = e

				// send message
				m := &Message{
					Domain: DomainDirectngm,
				}
				if e == pluginStateEnable {
					m.Command = CommDirectngmPlugEnabled
				} else if e == pluginStateDisable {
					m.Command = CommDirectngmPlugDisabled
				}
				jm, err := json.Marshal(m)
				if err != nil {
					log.Println(err)
					log.Println(m)
					return
				}
				writeMess(jm)
			}()

		case <-pl.quit:
			return
		}
	}
}

func handleTCPPlugin(c io.ReadWriteCloser, cv *CommentViewer) {
	defer cv.wg.Done()

	endc := make(chan bool, 1)

	cv.wg.Add(1)
	go func() {
		defer cv.wg.Done()
		select {
		// For quitting while receiving first init message.
		case <-cv.quit:
			err := c.Close()
			if err != nil {
				log.Println(err)
			}
		case iserr := <-endc:
			if iserr {
				err := c.Close()
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	dec := json.NewDecoder(c)

	m := new(Message)
	// It may stop here long time
	err := dec.Decode(m)
	if err != nil {
		log.Println(err)
		endc <- true
		return
	}
	if m.Domain != DomainDirect || m.Command != CommDirectNo {
		log.Println("send Direct.No message at first")
		endc <- true
		return
	}

	var ct CtDirectNo
	if err := json.Unmarshal(m.Content, &ct); err != nil {
		log.Println(err)
		endc <- true
		return
	}

	n := ct.No
	p, err := cv.Plugin(n)
	if err != nil {
		log.Panicln(err)
		endc <- true
		return
	}
	err = p.Open(c, !cv.Settings.PluginDisable[p.Name])
	if err != nil {
		log.Println(err)
		endc <- true
		return
	}
	log.Printf("loaded plugin : %s\n", p.Name)
	endc <- false
}

func handleSTDPlugin(p *plugin, cv *CommentViewer, path string) {
	defer cv.wg.Done()

	if len(p.Exec) < 1 {
		log.Printf("exec is not specified in plugin [%s]\n", p.Name)
		return
	}

	cmd := exec.Command(p.Exec[0], p.Exec[1:]...)
	cmd.Dir = path
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Println(err)
		return
	}
	needClose := true
	defer func() {
		if needClose {
			err = stdin.Close()
			if err != nil {
				log.Println(err)
			}
		}
	}()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		if needClose {
			err = stdout.Close()
			if err != nil {
				log.Println(err)
			}
		}
	}()
	err = cmd.Start()
	if err != nil {
		log.Println(err)
		return
	}

	c := &stdReadWriteCloser{stdout, stdin}
	err = p.Open(c, !cv.Settings.PluginDisable[p.Name])
	if err != nil {
		log.Println(err)
		return
	}
	needClose = false
	log.Println("loaded plugin : ", p.Name)
}

// Close closes opened plugin.
func (pl *plugin) Close() {
	pl.close()
	pl.wg.Wait()
}

func (pl *plugin) close() {
	select {
	case <-pl.quit:
	default:
		close(pl.quit)
	}
}

type stdReadWriteCloser struct {
	io.ReadCloser
	io.WriteCloser
}

func (rwc *stdReadWriteCloser) Close() error {
	errr := rwc.ReadCloser.Close()
	errw := rwc.WriteCloser.Close()
	if errr != nil {
		return errr
	}
	if errw != nil {
		return errw
	}
	return nil
}
