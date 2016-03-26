package nicolive

import (
	"bufio"
	"fmt"
	"html"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/xmlpath.v2"
)

const (
	numCommentConnectionRoutines = 2
	keepAliveDuration            = time.Minute
	postKeyDuration              = 10 * time.Second
)

// Comment is struct to hold a comment
type Comment struct {
	No        int
	Date      time.Time
	UserID    string
	Premium   int
	Anonymity bool
	Comment   string
	Mail      string
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

Main -> socket : dial
activate socket

note right
    returns no error
    even if failed to connect
end note

Main -> event : opened

create "receiveStream()" as rs
Main -> rs : go
activate rs
create "timer()" as tm
Main -> tm : go
activate tm


loop
...Wait for a message or closing socket...
rs -> event : comment,\nconnect,\nsubmit status
note left
    wait for a message
    even if connection error occured
end note
end

== Disconnecting ==
User -> Main : Disconnect
Main -> socket : close
deactivate socket

Main -> rs : termSig
destroy rs
Main -> tm : termSig
destroy tm

Main -> event : disconnect
@enduml
*/
type CommentConnection struct {
	IsConnected bool
	liveWaku    *LiveWaku
	socket      net.Conn
	ticket      string
	svrTm       time.Time
	svrTmD      time.Duration
	openTm      time.Time

	lastBlock int

	rw bufio.ReadWriter

	keepAliveTmr *time.Timer
	postKeyTmr   *time.Timer
	wmu          sync.Mutex
	termc        chan bool
}

// NewCommentConnection returns a pointer to new CommentConnection
func NewCommentConnection(l *LiveWaku) *CommentConnection {
	kat := time.NewTimer(keepAliveDuration)
	kat.Stop()
	pkt := time.NewTimer(postKeyDuration)
	pkt.Stop()

	return &CommentConnection{
		liveWaku:     l,
		termc:        make(chan bool),
		keepAliveTmr: kat,
		postKeyTmr:   pkt,
	}
}

func (cc *CommentConnection) open() {
	var err error
	Logger.Println("CommentConnection opening")

	addrport := fmt.Sprintf("%s:%s",
		cc.liveWaku.CommentServer.Addr,
		cc.liveWaku.CommentServer.Port)

	cc.wmu.Lock()

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

	cc.wmu.Unlock()

	cc.openTm = time.Now()
	EvReceiver.Proceed(&Event{EventString: "comment connection opened"})
}

// Connect Connect to nicolive and start receiving comment
func (cc *CommentConnection) Connect() NicoError {
	if cc.IsConnected {
		return NicoErr(NicoErrOther, "already connected", "")
	}
	cc.IsConnected = true

	cc.open()
	cc.keepAliveTmr.Reset(keepAliveDuration)

	go cc.receiveStream()
	go cc.timer()

	return nil
}

func (cc *CommentConnection) receiveStream() {
	for {
		select {
		case <-cc.termc:
			return
		default:
			commxml, err := cc.rw.ReadString('\x00')
			if err != nil {
				go cc.Disconnect()
				Logger.Println(NicoErrFromStdErr(err))
				<-cc.termc
				return
			}

			// strip null char
			commxml = commxml[:len(commxml)-1]

			fmt.Println(commxml)

			commxmlReader := strings.NewReader(commxml)
			rt, err := xmlpath.Parse(commxmlReader)
			if err != nil {
				Logger.Println(NicoErrFromStdErr(err))
				continue
			}

			if strings.HasPrefix(commxml, "<thread ") {
				if v, ok := xmlpath.MustCompile("/thread/@last_res").String(rt); ok {
					lbl, _ := strconv.Atoi(v)
					cc.lastBlock = lbl / 10
				}
				if v, ok := xmlpath.MustCompile("/thread/@ticket").String(rt); ok {
					cc.ticket = v
				}
				if v, ok := xmlpath.MustCompile("/thread/@server_time").String(rt); ok {
					i, _ := strconv.ParseInt(v, 10, 64)
					cc.svrTm = time.Unix(i, 0)
					cc.svrTmD = time.Now().Sub(cc.svrTm)
				}

				// immediately update postkey and start the timer
				cc.postKeyTmr.Reset(0)

				continue
			}
			if strings.HasPrefix(commxml, "<chat_result ") {
				if v, ok := xmlpath.MustCompile("/chat_result/@status").String(rt); ok {
					if v != "0" {
						Logger.Println("comment submit error (chat_result status): " + v)
						continue
					}
					EvReceiver.Proceed(&Event{EventString: "comment submitted"})
				}
				continue
			}
			if strings.HasPrefix(commxml, "<chat ") {
				var comment Comment

				if v, ok := xmlpath.MustCompile("/chat").String(rt); ok {
					comment.Comment = html.UnescapeString(v)
				}
				if v, ok := xmlpath.MustCompile("/chat/@no").String(rt); ok {
					comment.No, _ = strconv.Atoi(v)
				}
				if v, ok := xmlpath.MustCompile("/chat/@date").String(rt); ok {
					i, _ := strconv.ParseInt(v, 10, 64)
					comment.Date = time.Unix(i, 0)
				}
				if v, ok := xmlpath.MustCompile("/chat/@mail").String(rt); ok {
					comment.Mail = v
				}
				if v, ok := xmlpath.MustCompile("/chat/@user_id").String(rt); ok {
					comment.UserID = v
				}
				if v, ok := xmlpath.MustCompile("/chat/@premium").String(rt); ok {
					comment.Premium, _ = strconv.Atoi(v)
				}
				if v, ok := xmlpath.MustCompile("/chat/@anonymity").String(rt); ok {
					comment.Anonymity, _ = strconv.ParseBool(v)
				}

				blk := comment.No / 10
				if blk > cc.lastBlock {
					cc.lastBlock = blk
					cc.postKeyTmr.Reset(0)
				}

				EvReceiver.Proceed(&Event{
					EventString: "comment",
					Content:     comment,
				})

				if comment.Comment == "/disconnect" &&
					(comment.Premium>>1)%2 == 1 {

					go cc.Disconnect()
					<-cc.termc
					return
				}
				continue
			}
		}
	}
}

func (cc *CommentConnection) timer() {
	for {
		select {
		case <-cc.termc:
			return
		case <-cc.keepAliveTmr.C:
			cc.keepAliveTmr.Reset(keepAliveDuration)
			cc.wmu.Lock()
			err := cc.rw.WriteByte(0)
			if err == nil {
				err = cc.rw.Flush()
			}
			cc.wmu.Unlock()
			if err != nil {
				go cc.Disconnect()
				Logger.Println(NicoErrFromStdErr(err))
				<-cc.termc
				return
			}
		case <-cc.postKeyTmr.C:
			cc.postKeyTmr.Reset(postKeyDuration)
			nerr := cc.liveWaku.FetchPostKey(cc.lastBlock)
			if nerr != nil {
				Logger.Println(nerr)
			}
		}
	}
}

// Disconnect close and disconnect
// terminate all goroutines and wait to exit
func (cc *CommentConnection) Disconnect() NicoError {
	if !cc.IsConnected {
		return NicoErr(NicoErrOther, "not connected yet", "")
	}

	cc.keepAliveTmr.Stop()
	cc.postKeyTmr.Stop()
	cc.socket.Close()

	for i := 0; i < numCommentConnectionRoutines; i++ {
		cc.termc <- true
	}

	cc.IsConnected = false
	Logger.Println("CommentConnection disconnected")

	return nil
}
