package viewer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"sync"
	"time"

	"github.com/diginatu/nagome/api"
	"github.com/diginatu/nagome/services/utils"
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

// A Plugin is a Nagome plugin.
type Plugin struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Version     string      `yaml:"version"`
	Author      string      `yaml:"author"`
	Method      string      `yaml:"method"`
	Exec        []string    `yaml:"exec"`
	Nagomever   string      `yaml:"nagomever"`
	Subscribe   []string    `yaml:"subscribe"`
	No          int         `yaml:"-"`
	GetState    pluginState `yaml:"-"` // Don't change directly
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
func newPlugin(cv *CommentViewer) *Plugin {
	return &Plugin{
		No:         -1,
		quit:       make(chan struct{}),
		setStateCh: make(chan pluginState),
		writec:     make(chan []byte, pluginEachMessageChanSize),
		cv:         cv,
	}
}

// API returns instance of API representation of the Plugin
func (pl *Plugin) API() *api.Plugin {
	return &api.Plugin{
		Name:        pl.Name,
		Description: pl.Description,
		Version:     pl.Version,
		Author:      pl.Author,
		Method:      pl.Method,
		Subscribe:   pl.Subscribe,
		No:          pl.No,
		State:       int(pl.GetState),
	}
}

// NewPluginsAPI creates new instance of API representation of the Plugins
func NewPluginsAPI(pls []*Plugin) []*api.Plugin {
	apiPls := make([]*api.Plugin, len(pls))

	for i, pl := range pls {
		apiPls[i] = pl.API()
	}

	return apiPls
}

// Open opens connection and start processing.
func (pl *Plugin) Open(rwc io.ReadWriteCloser, enable bool) error {
	pl.stateMu.Lock()
	defer pl.stateMu.Unlock()

	if pl.No == -1 {
		return fmt.Errorf("plugin \"%s\" is not initialized (add to CommentViewer)", pl.Name)
	}
	if pl.Name == "" {
		return fmt.Errorf("plugin \"%s\" no name is set", pl.Name)
	}
	if rwc == nil {
		return fmt.Errorf("given rw is nil")
	}
	if pl.GetState != pluginStateClose {
		return fmt.Errorf("already opened")
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

// SetState sets state of the plugin.
func (pl *Plugin) SetState(enable bool) {
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

// WriteMess writes a Nagome message into the plugin.
func (pl *Plugin) WriteMess(m *api.Message) (fail bool) {
	jm, err := json.Marshal(m)
	if err != nil {
		pl.cv.cli.log.Println(err)
		pl.cv.cli.log.Println(m)
		return
	}
	return pl.Write(jm)
}

func (pl *Plugin) Write(p []byte) (fail bool) {
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

// IsSubscribe returns whether the plugin subscribes given Domain.
func (pl *Plugin) IsSubscribe(pln string) bool {
	f := false
	for _, d := range pl.Subscribe {
		if d == pln {
			f = true
			break
		}
	}
	return f
}

// Load loads from file and set values.
func (pl *Plugin) Load(filePath string) error {
	d, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(d, pl)
}

// Save saves the plugin into a file with given name.
func (pl *Plugin) Save(filePath string) error {
	d, err := yaml.Marshal(pl)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, d, 0600)
}

// IsMain returns whether the plugin is main plugin.
func (pl *Plugin) IsMain() bool {
	return pl.No == 0
}

func (pl *Plugin) evRoutine() {
	defer pl.wg.Done()
	defer func() {
		err := pl.rwc.Close()
		if err != nil {
			pl.cv.cli.log.Println(err)
		}
		pl.stateMu.Lock()
		pl.GetState = pluginStateClose
		pl.stateMu.Unlock()
	}()
	defer pl.cv.cli.log.Printf("plugin [%s] is closing", pl.Name)

	// Run decoder.  It puts a message into "mes".
	dec := json.NewDecoder(pl.rwc)
	mes := make(chan (*api.Message))
	pl.wg.Add(1)
	go func() {
		defer pl.wg.Done()
		for {
			m := new(api.Message)
			err := dec.Decode(m)
			if err != nil {
				select {
				// ignore if quitting
				case <-pl.quit:
				default:
					if err != io.EOF {
						utils.EmitEvNewNotification(api.CtUINotificationTypeInfo,
							"plugin disconnected",
							fmt.Sprintf("plugin [%s] : connection disconnected", pl.Name),
							pl.cv.Evch, pl.cv.cli.log)
						pl.cv.cli.log.Println(err)
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
			pl.cv.cli.log.Println(err)
			utils.EmitEvNewNotification(api.CtUINotificationTypeInfo,
				"plugin", "failed to write a message : "+pl.Name,
				pl.cv.Evch, pl.cv.cli.log)
			// quit if UI plugin disconnect
			if pl.IsMain() {
				pl.cv.Quit()
			} else {
				pl.close()
			}
		}
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
			m.Plgno = pl.No
			pl.cv.cli.log.Printf("plugin message [%s] : %v", pl.Name, m)
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
				pl.cv.cli.log.Println(err)
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
				m := &api.Message{
					Domain: api.DomainDirectngm,
				}
				if e == pluginStateEnable {
					m.Command = api.CommDirectngmPlugEnabled
				} else if e == pluginStateDisable {
					m.Command = api.CommDirectngmPlugDisabled
				}
				jm, err := json.Marshal(m)
				if err != nil {
					pl.cv.cli.log.Println(err)
					pl.cv.cli.log.Println(m)
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
				cv.cli.log.Println(err)
			}
		case iserr := <-endc:
			if iserr {
				err := c.Close()
				if err != nil {
					cv.cli.log.Println(err)
				}
			}
		}
	}()

	dec := json.NewDecoder(c)

	m := new(api.Message)
	// It may stop here long time
	err := dec.Decode(m)
	if err != nil {
		cv.cli.log.Println(err)
		endc <- true
		return
	}
	if m.Domain != api.DomainDirect || m.Command != api.CommDirectNo {
		cv.cli.log.Println("send Direct.No message at first")
		endc <- true
		return
	}

	var ct api.CtDirectNo
	if err := json.Unmarshal(m.Content, &ct); err != nil {
		cv.cli.log.Println(err)
		endc <- true
		return
	}

	n := ct.No
	p, err := cv.Plugin(n)
	if err != nil {
		cv.cli.log.Panicln(err)
		endc <- true
		return
	}
	err = p.Open(c, !cv.Settings.PluginDisable[p.Name])
	if err != nil {
		cv.cli.log.Println(err)
		endc <- true
		return
	}
	cv.cli.log.Printf("loaded plugin : %s\n", p.Name)
	endc <- false
}

func handleSTDPlugin(p *Plugin, cv *CommentViewer, path string) {
	defer cv.wg.Done()

	if len(p.Exec) < 1 {
		cv.cli.log.Printf("exec is not specified in plugin [%s]\n", p.Name)
		return
	}

	cmd := exec.Command(p.Exec[0], p.Exec[1:]...)
	cmd.Dir = path
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cv.cli.log.Println(err)
		return
	}
	needClose := true
	defer func() {
		if needClose {
			err = stdin.Close()
			if err != nil {
				cv.cli.log.Println(err)
			}
		}
	}()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cv.cli.log.Println(err)
		return
	}
	defer func() {
		if needClose {
			err = stdout.Close()
			if err != nil {
				cv.cli.log.Println(err)
			}
		}
	}()
	err = cmd.Start()
	if err != nil {
		cv.cli.log.Println(err)
		return
	}

	c := &stdReadWriteCloser{stdout, stdin}
	err = p.Open(c, !cv.Settings.PluginDisable[p.Name])
	if err != nil {
		cv.cli.log.Println(err)
		return
	}
	needClose = false
	cv.cli.log.Println("loaded plugin : ", p.Name)
}

// Close closes opened plugin.
func (pl *Plugin) Close() {
	pl.close()
	pl.wg.Wait()
}

func (pl *Plugin) close() {
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
