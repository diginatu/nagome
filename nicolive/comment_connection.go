package nicolive

import (
	"bufio"
	"fmt"
	"html"
	"io/ioutil"
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
	No          int
	Date        time.Time
	UserID      string
	IsPremium   bool
	IsCommand   bool
	IsStaff     bool
	IsAnonymity bool
	Comment     string
	Mail        string
	Locale      string
	Score       int
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
//send.append('\0');

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

	lv     *LiveWaku
	sock   net.Conn
	ticket string
	svrDu  time.Duration
	block  int

	rw           bufio.ReadWriter
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
		lv:           l,
		termc:        make(chan bool),
		keepAliveTmr: kat,
		postKeyTmr:   pkt,
	}
}

func (cc *CommentConnection) open() {
	var err error
	Logger.Println("CommentConnection opening")

	addrport := fmt.Sprintf("%s:%s",
		cc.lv.CommentServer.Addr,
		cc.lv.CommentServer.Port)

	cc.wmu.Lock()

	cc.sock, err = net.Dial("tcp", addrport)
	if err != nil {
		Logger.Println(NicoErrFromStdErr(err))
		return
	}

	cc.rw = bufio.ReadWriter{
		Reader: bufio.NewReader(cc.sock),
		Writer: bufio.NewWriter(cc.sock),
	}

	fmt.Fprintf(cc.rw,
		"<thread thread=\"%s\" res_from=\"-1000\" version=\"20061206\" />\x00",
		cc.lv.CommentServer.Thread)
	err = cc.rw.Flush()
	if err != nil {
		Logger.Println(NicoErrFromStdErr(err))
		return
	}

	cc.wmu.Unlock()

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

			commxmlr := strings.NewReader(commxml)
			rt, err := xmlpath.Parse(commxmlr)
			if err != nil {
				Logger.Println(NicoErrFromStdErr(err))
				continue
			}

			if strings.HasPrefix(commxml, "<thread ") {
				if v, ok := xmlpath.MustCompile("/thread/@last_res").String(rt); ok {
					lbl, _ := strconv.Atoi(v)
					cc.block = lbl / 10
				}
				if v, ok := xmlpath.MustCompile("/thread/@ticket").String(rt); ok {
					cc.ticket = v
				}
				if v, ok := xmlpath.MustCompile("/thread/@server_time").String(rt); ok {
					i, _ := strconv.ParseInt(v, 10, 64)
					cc.svrDu = time.Unix(i, 0).Sub(time.Now())
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
				if d, ok := xmlpath.MustCompile("/chat/@date").String(rt); ok {
					di, _ := strconv.ParseInt(d, 10, 64)
					var udi int64
					if ud, ok := xmlpath.MustCompile("/chat/@date_usec").String(rt); ok {
						udi, _ = strconv.ParseInt(ud, 10, 64)
					}
					comment.Date = time.Unix(di, udi*1000)
				}
				if v, ok := xmlpath.MustCompile("/chat/@mail").String(rt); ok {
					comment.Mail = v
				}
				if v, ok := xmlpath.MustCompile("/chat/@user_id").String(rt); ok {
					comment.UserID = v
				}
				if v, ok := xmlpath.MustCompile("/chat/@premium").String(rt); ok {
					i, _ := strconv.Atoi(v)
					comment.IsPremium = i%2 == 1
					i >>= 1
					comment.IsCommand = i%2 == 1
					i >>= 1
					comment.IsStaff = i%2 == 1
				}
				if v, ok := xmlpath.MustCompile("/chat/@anonymity").String(rt); ok {
					comment.IsAnonymity, _ = strconv.ParseBool(v)
				}
				if v, ok := xmlpath.MustCompile("/chat/@locale").String(rt); ok {
					comment.Locale = v
				}
				if v, ok := xmlpath.MustCompile("/chat/@score").String(rt); ok {
					comment.Score, _ = strconv.Atoi(v)
				}

				blk := comment.No / 10
				if blk > cc.block {
					cc.block = blk
					cc.postKeyTmr.Reset(0)
				}

				EvReceiver.Proceed(&Event{
					EventString: "comment",
					Content:     comment,
				})

				if comment.IsCommand && comment.Comment == "/disconnect" {
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
			nerr := cc.FetchPostKey()
			if nerr != nil {
				Logger.Println(nerr)
			}
		}
	}
}

// FetchPostKey gets postkey using getpostkey API
func (cc *CommentConnection) FetchPostKey() NicoError {
	if cc.lv.Account == nil {
		return NicoErr(NicoErrOther, "no account",
			"LiveWaku does not have an account")
	}
	if cc.lv.BroadID == "" {
		return NicoErr(NicoErrOther, "no BroadID",
			"BroadID is not set")
	}

	c, nicoerr := NewNicoClient(cc.lv.Account)
	if nicoerr != nil {
		return nicoerr
	}

	url := fmt.Sprintf(
		"http://live.nicovideo.jp/api/getpostkey?thread=%s&block_no=%s",
		cc.lv.CommentServer.Thread, cc.block)
	res, err := c.Get(url)
	if err != nil {
		return NicoErrFromStdErr(err)
	}
	defer res.Body.Close()

	allb, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return NicoErrFromStdErr(err)
	}

	pk := string(allb[8:])
	if pk == "" {
		return NicoErr(NicoErrOther, "failed to get postkey", "")
	}

	cc.lv.PostKey = pk

	return nil
}

// SendComment sends comment to current comment connection
func (cc *CommentConnection) SendComment(text string, iyayo bool) {
	if !cc.IsConnected {
		Logger.Println(NicoErrOther, "not connected", "")
		return
	}
	if cc.lv.PostKey == "" {
		Logger.Println(NicoErrOther, "no postkey", "")
		return
	}
	if text == "" {
		Logger.Println(NicoErrOther, "empty text", "")
		return
	}

	vpos := 100 *
		(time.Now().Add(cc.svrDu).Unix() - cc.lv.Stream.OpenTime.Unix())

	var iyayos string
	if iyayo {
		iyayos = " mail=\"184\""
	}
	var prems string
	if cc.lv.User.IsPremium {
		prems = " premium=\"1\""
	}

	sdcomm := fmt.Sprintf("<chat thread=\"%s\" ticket=\"%s\" "+
		"vpos=\"%d\" postkey=\"%s\"%s user_id=\"%s\"%s>%s</chat>\x00",
		cc.lv.CommentServer.Thread,
		cc.ticket,
		vpos,
		cc.lv.PostKey,
		iyayos,
		cc.lv.User.UserID,
		prems,
		html.EscapeString(text))

	fmt.Println(sdcomm)

	cc.wmu.Lock()
	fmt.Fprint(cc.rw, sdcomm)
	err := cc.rw.Flush()
	if err != nil {
		Logger.Println(NicoErrFromStdErr(err))
		return
	}
	cc.wmu.Unlock()
	cc.keepAliveTmr.Reset(keepAliveDuration)
}

// Disconnect close and disconnect
// terminate all goroutines and wait to exit
func (cc *CommentConnection) Disconnect() NicoError {
	if !cc.IsConnected {
		return NicoErr(NicoErrOther, "not connected yet", "")
	}

	cc.keepAliveTmr.Stop()
	cc.postKeyTmr.Stop()
	cc.sock.Close()

	for i := 0; i < numCommentConnectionRoutines; i++ {
		cc.termc <- true
	}

	cc.IsConnected = false
	Logger.Println("CommentConnection disconnected")

	return nil
}
