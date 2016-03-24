package nicolive

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/xmlpath.v2"
)

const (
	numCommentConnectionGoRoutines = 2
)

// Comment is struct to hold a comment
type Comment struct {
	No        int
	Date      time.Time
	UserID    string
	Premium   int
	Anonymity bool
	Comment   string
}

// CommentConnection is a struct to manage sending/receiving comments.
// This struct automatically submits NULL character to reserve connection and
// get the PostKey, which is necessary for sending comments.
// liveWaku should have connection information which is able to get by fetchInformation()
/*
@startuml
title "Comment connection - Sequence Diagram"
actor User

== Connecting ==
User -> Main : Connect
create "receiveStream()" as rs
Main -> rs : lock and go
create "keepAlive()" as kp
Main -> kp : lock and go

create "open()" as open
Main -> open : go
activate open

open -> socket : dial
activate socket

open -> rs : unlock
destroy open
note right
    open returns no error
    even if failed to connect
end note
activate rs
open -> kp : unlock
activate kp

open -> event : open


loop
...Wait for a comment or closing socket...
rs -> event : comment
note left
    wait for a comment
    even if connection error occured
end note
end

== Disconnecting ==
User -> Main : Disconnect
Main -> socket : close
deactivate socket

Main -> rs : termSig
destroy rs
Main -> kp : termSig
destroy kp
Main -> Main : close

Main -> event : disconnect
@enduml
*/
type CommentConnection struct {
	IsConnected bool
	liveWaku    *LiveWaku
	socket      net.Conn
	ticket      string
	svrTime     time.Time
	svrTimeDiff time.Duration

	lastBlock int

	rw bufio.ReadWriter

	rMutex  sync.Mutex
	wMutex  sync.Mutex
	termSig chan bool
}

// NewCommentConnection returns a pointer to new CommentConnection
func NewCommentConnection(l *LiveWaku) *CommentConnection {
	return &CommentConnection{
		liveWaku: l,
		termSig:  make(chan bool),
	}
}

func (cc *CommentConnection) open() {
	var err error
	Logger.Println("CommentConnection opening")

	addrport := fmt.Sprintf("%s:%s",
		cc.liveWaku.CommentServer.Addr,
		cc.liveWaku.CommentServer.Port)

	cc.socket, err = net.Dial("tcp", addrport)
	if err != nil {
		Logger.Println(NicoErrFromStdErr(err))
		return
	}

	cc.rw = bufio.ReadWriter{
		Reader: bufio.NewReader(cc.socket),
		Writer: bufio.NewWriter(cc.socket),
	}

	fmt.Fprintf(cc.rw,
		"<thread thread=\"%s\" res_from=\"-1000\" version=\"20061206\" />\x00",
		cc.liveWaku.CommentServer.Thread)
	err = cc.rw.Flush()
	if err != nil {
		Logger.Println(NicoErrFromStdErr(err))
		return
	}
}

// Connect Connect to nicolive and start receiving comment
func (cc *CommentConnection) Connect() NicoError {
	if cc.IsConnected {
		return NicoErr(NicoErrOther, "already connected", "")
	}
	cc.IsConnected = true

	cc.wMutex.Lock()
	cc.rMutex.Lock()

	go func() {
		cc.open()

		cc.wMutex.Unlock()
		cc.rMutex.Unlock()
	}()

	go cc.receiveStream()
	go cc.keepAlive()

	return nil
}

func (cc *CommentConnection) receiveStream() {
	for {
		select {
		case <-cc.termSig:
			return
		default:
			cc.rMutex.Lock()
			commxml, err := cc.rw.ReadString('\x00')
			cc.rMutex.Unlock()
			if err != nil {
				Logger.Println(NicoErrFromStdErr(err))
				continue
			}

			// strip null char
			commxml = commxml[0 : len(commxml)-1]

			fmt.Println(commxml)

			if strings.HasPrefix(commxml, "<thread ") {
				commxmlReader := strings.NewReader(commxml)

				root, err := xmlpath.Parse(commxmlReader)
				if err != nil {
					Logger.Println(NicoErrFromStdErr(err))
					continue
				}

				if v, ok := xmlpath.MustCompile("/thread/@last_res").String(root); ok {
					lbl, _ := strconv.Atoi(v)
					cc.lastBlock = lbl / 10
				}
				if v, ok := xmlpath.MustCompile("/thread/@ticket").String(root); ok {
					cc.ticket = v
				}
				if v, ok := xmlpath.MustCompile("/thread/@server_time").String(root); ok {
					i, _ := strconv.ParseInt(v, 10, 64)
					cc.svrTime = time.Unix(i, 0)
					cc.svrTimeDiff = time.Now().Sub(cc.svrTime)
				}

				//livewaku->fetchPostKey(lastBlockNum, userSession);
				//postkeyTimer->start(10000);

				continue
			}
			if strings.HasPrefix(commxml, "<chat ") {
				commxmlReader := strings.NewReader(commxml)
				var comment Comment

				root, err := xmlpath.Parse(commxmlReader)
				if err != nil {
					Logger.Println(NicoErrFromStdErr(err))
					continue
				}

				if v, ok := xmlpath.MustCompile("/chat").String(root); ok {
					comment.Comment = v
				}
				if v, ok := xmlpath.MustCompile("/chat/@no").String(root); ok {
					comment.No, _ = strconv.Atoi(v)
				}
				if v, ok := xmlpath.MustCompile("/chat/@premium").String(root); ok {
					comment.Premium, _ = strconv.Atoi(v)
				}
				if v, ok := xmlpath.MustCompile("/chat/@date").String(root); ok {
					i, _ := strconv.ParseInt(v, 10, 64)
					comment.Date = time.Unix(i, 0)
				}
				if v, ok := xmlpath.MustCompile("/chat/@anonymity").String(root); ok {
					comment.Anonymity, _ = strconv.ParseBool(v)
				}

				EvReceiver.Proceed(&Event{
					EventString: "comment",
					Content:     comment,
				})
				continue
			}
		}
	}
}

func (cc *CommentConnection) keepAlive() {
	tick := time.Tick(time.Minute)
	for {
		select {
		case <-cc.termSig:
			return
		case <-tick:
			cc.wMutex.Lock()
			err := cc.rw.WriteByte(0)
			if err == nil {
				err = cc.rw.Flush()
			}
			cc.wMutex.Unlock()
			if err != nil {
				Logger.Println(NicoErrFromStdErr(err))
				continue
			}
		}
	}
}

// Close closes the connection
// do not end go routines
func (cc *CommentConnection) close() {
	cc.socket.Close()
}

// Disconnect close and disconnect
// terminate all goroutines and wait to exit
func (cc *CommentConnection) Disconnect() NicoError {
	if !cc.IsConnected {
		return NicoErr(NicoErrOther, "not connected yet", "")
	}

	cc.close()
	for i := 0; i < numCommentConnectionGoRoutines; i++ {
		cc.termSig <- true
	}

	cc.IsConnected = false
	Logger.Println("CommentConnection disconnected")

	return nil
}
