package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"testing"

	"github.com/diginatu/nagome/nicolive"
)

type testRwc struct {
	r io.Reader
	w io.Writer
	c io.Closer
	t *testing.T
}

func (rwc *testRwc) Read(p []byte) (n int, err error) {
	rwc.t.Logf("rwc read : %s\n", p)
	return rwc.r.Read(p)
}

func (rwc *testRwc) Write(p []byte) (n int, err error) {
	rwc.t.Logf("rwc write : %s\n", p)
	return rwc.w.Write(p)
}

func (rwc *testRwc) Close() error {
	rwc.t.Logf("rwc closed\n")
	return rwc.c.Close()
}

func TestPluginClose(t *testing.T) {
	ac := &nicolive.Account{Mail: "mail", Pass: "pass", Usersession: "session"}
	cv := NewCommentViewer(ac, "0")
	p := newPlugin(cv)
	p.Name = "main"
	p.Description = "main plugin"
	p.Version = "0.0"
	p.Method = pluginMethodTCP
	p.Depends = []string{DomainNagome, DomainComment, DomainUI}

	// assume be added as main plugin.
	p.No = 1

	//b := new(bytes.Buffer)
	pr, _ := io.Pipe()
	rwc := &testRwc{pr, ioutil.Discard, pr, t}

	wait := make(chan struct{})

	// handle TCP routine
	go func() {
		err := p.Open(rwc)
		if err != nil {
			t.Fatal(err)
		}
		wait <- struct{}{}
	}()

	<-wait

	p.Close()
}

func TestPluginErrorConnection(t *testing.T) {
	ac := &nicolive.Account{Mail: "mail", Pass: "pass", Usersession: "session"}
	cv := NewCommentViewer(ac, "0")
	p := newPlugin(cv)
	p.Name = "normal"
	p.Description = "normal plugin"
	p.Version = "0.0"
	p.Method = pluginMethodTCP
	p.Depends = []string{DomainNagome, DomainComment, DomainUI}

	// assume be added as normal plugin.
	p.No = 2

	//b := new(bytes.Buffer)
	pr, _ := io.Pipe()
	rwc := &testRwc{pr, ioutil.Discard, pr, t}

	wait := make(chan struct{})

	// handle TCP routine
	go func() {
		err := p.Open(rwc)
		if err != nil {
			t.Fatal(err)
		}
		wait <- struct{}{}
	}()

	<-wait

	pr.Close()
	p.wg.Wait()
}

func TestPluginWrite(t *testing.T) {
	ac := &nicolive.Account{Mail: "mail", Pass: "pass", Usersession: "session"}
	cv := NewCommentViewer(ac, "0")
	p := newPlugin(cv)
	p.Name = "normal"
	p.Description = "normal plugin"
	p.Version = "0.0"
	p.Method = pluginMethodTCP
	p.Depends = []string{DomainNagome, DomainComment, DomainUI}

	// assume be added as normal plugin.
	p.No = 2

	// no effect before opening
	p.WriteMess(&Message{
		Domain:  "fail",
		Command: "This message should not be send.",
	})

	//b := new(bytes.Buffer)
	pr, _ := io.Pipe()
	pwreader, pw := io.Pipe()
	rwc := &testRwc{pr, pw, pr, t}

	// handle TCP routine (to test race)
	wait := make(chan struct{})
	go func() {
		err := p.Open(rwc)
		if err != nil {
			t.Fatal(err)
		}
		wait <- struct{}{}
	}()
	<-wait

	p.WriteMess(&Message{
		Domain:  "ok",
		Command: "This message must be send.",
	})

	dec := json.NewDecoder(pwreader)
	for {
		m := new(Message)
		err := dec.Decode(m)
		if err != nil {
			t.Fatal(err)
		}
		if m.Domain == "fail" {
			t.Fatal(m.Command)
		}
		if m.Domain == "ok" {
			break
		}
	}

	p.Close()
}
