package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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

func pluginTCPServer(cv *commentViewer, wg *sync.WaitGroup) {
	defer wg.Done()
	adr, err := net.ResolveTCPAddr("tcp", ":"+tcpPort)
	if err != nil {
		Logger.Panicln(err)
	}
	l, err := net.ListenTCP("tcp", adr)
	if err != nil {
		Logger.Panicln(err)
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
				Logger.Println(err)
				continue
			}
			wg.Add(1)
			go handleTCPPlugin(conn, cv, wg)
		case <-cv.Quit:
			return
		}
	}
}

func handleTCPPlugin(c net.Conn, cv *commentViewer, wg *sync.WaitGroup) {
	defer wg.Done()
	defer c.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
	for {
		select {
		default:
			c.SetReadDeadline(time.Now().Add(time.Second))
			b, _, err := rw.ReadLine()
			nerr, ok := err.(net.Error)
			if ok && nerr.Timeout() && nerr.Temporary() {
				continue
			}
			if err != nil {
				return
			}
			Logger.Println(b)
		case <-cv.Quit:
			return
		}
	}
}
