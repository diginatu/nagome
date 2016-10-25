package nicolive

import (
	"bufio"
	"context"
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
	heartbeatDuration            = 90 * time.Second
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
// This struct automatically Keeping alive and get the PostKey, which is necessary for sending comments.
type CommentConnection struct {
	ConnectedTm time.Time // The time when this connection started

	lv     *LiveWaku
	conn   net.Conn
	ticket string
	svrDu  time.Duration
	block  int

	rw            bufio.ReadWriter
	postKeyTmr    *time.Timer
	heartbeatTmr  *time.Timer
	wmu           sync.Mutex
	cancel        context.CancelFunc
	ev            EventReceiver
	disconnecting bool
	wg            sync.WaitGroup
}

// CommentConnect connects to nicolive and start receiving comment.
// liveWaku should have connection information which is able to get by fetchInformation()
func CommentConnect(ctx context.Context, lv *LiveWaku, ev EventReceiver) (*CommentConnection, Error) {
	kat := time.NewTimer(keepAliveDuration)
	kat.Stop()
	pkt := time.NewTimer(postKeyDuration)
	pkt.Stop()
	hbt := time.NewTimer(heartbeatDuration)
	hbt.Stop()

	if ev == nil {
		ev = &defaultEventReceiver{}
	}

	cc := &CommentConnection{
		lv:           lv,
		postKeyTmr:   pkt,
		heartbeatTmr: hbt,
		ev:           ev,
	}

	cctx, cancel := context.WithCancel(context.Background())
	cc.cancel = cancel

	nerr := cc.open(ctx)
	if nerr != nil {
		return nil, nerr
	}

	cc.wg.Add(2)
	go cc.receiveStream(cctx)
	go cc.timer(cctx)

	return cc, nil
}

func (cc *CommentConnection) open(ctx context.Context) Error {
	if cc.lv.Account == nil {
		return MakeError(ErrOther, "nil account in LiveWaku")
	}
	if cc.lv.BroadID == "" {
		return MakeError(ErrOther, "BroadID is not set")
	}

	var err error

	addrport := net.JoinHostPort(cc.lv.CommentServer.Addr, cc.lv.CommentServer.Port)

	d := &net.Dialer{
		KeepAlive: keepAliveDuration,
	}

	cc.conn, err = d.DialContext(ctx, "tcp", addrport)
	if err != nil {
		return ErrFromStdErr(err)
	}

	cc.rw = bufio.ReadWriter{
		Reader: bufio.NewReader(cc.conn),
		Writer: bufio.NewWriter(cc.conn),
	}

	_, err = fmt.Fprintf(cc.rw,
		"<thread thread=\"%s\" res_from=\"-1000\" version=\"20061206\" />\x00",
		cc.lv.CommentServer.Thread)
	if err != nil {
		return ErrFromStdErr(err)
	}
	err = cc.rw.Flush()
	if err != nil {
		return ErrFromStdErr(err)
	}

	return nil
}

func (cc *CommentConnection) receiveStream(ctx context.Context) {
	defer cc.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			commxml, err := cc.rw.ReadString('\x00')
			if err != nil {
				nerr, ok := err.(net.Error)
				if ok && nerr.Temporary() {
					continue
				}

				if cc.disconnecting {
					return
				}

				cc.ev.ProceedNicoEvent(&Event{
					Type:    EventTypeErr,
					Content: ErrFromStdErr(err),
				})
				go cc.Disconnect()
				return
			}

			// strip null char
			commxml = commxml[:len(commxml)-1]

			commxmlr := strings.NewReader(commxml)
			rt, err := xmlpath.Parse(commxmlr)
			if err != nil {
				cc.ev.ProceedNicoEvent(&Event{
					Type:    EventTypeErr,
					Content: ErrFromStdErr(err),
				})
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
					cc.ConnectedTm = time.Unix(i, 0)
					cc.svrDu = cc.ConnectedTm.Sub(time.Now())
				}

				// immediately update postkey and start the timer
				cc.postKeyTmr.Reset(0)
				cc.heartbeatTmr.Reset(0)

				cc.ev.ProceedNicoEvent(&Event{
					Type:    EventTypeOpen,
					Content: nil,
				})

				continue
			}
			if strings.HasPrefix(commxml, "<chat_result ") {
				if v, ok := xmlpath.MustCompile("/chat_result/@status").String(rt); ok {
					if v != "0" {
						cc.ev.ProceedNicoEvent(&Event{
							Type:    EventTypeErr,
							Content: MakeError(ErrSendComment, "comment send error. chat_result status : "+v),
						})
						continue
					}
					cc.ev.ProceedNicoEvent(&Event{
						Type:    EventTypeSend,
						Content: nil,
					})
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

				cc.ev.ProceedNicoEvent(&Event{
					Type:    EventTypeGot,
					Content: comment,
				})

				if comment.IsCommand && comment.Comment == "/disconnect" {
					go cc.Disconnect()
					cc.ev.ProceedNicoEvent(&Event{
						Type:    EventTypeWakuEnd,
						Content: *cc.lv,
					})
					return
				}
				continue
			}

			cc.ev.ProceedNicoEvent(&Event{
				Type:    EventTypeErr,
				Content: MakeError(ErrSendComment, "unknown stream : "+commxml),
			})
		}
	}
}

func (cc *CommentConnection) timer(ctx context.Context) {
	defer cc.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-cc.postKeyTmr.C:
			cc.postKeyTmr.Reset(postKeyDuration)
			nerr := cc.FetchPostKey()
			if nerr != nil {
				cc.ev.ProceedNicoEvent(&Event{
					Type:    EventTypeErr,
					Content: nerr,
				})
				continue
			}
		case <-cc.heartbeatTmr.C:
			cc.heartbeatTmr.Reset(heartbeatDuration)
			hbv, nerr := cc.lv.FetchHeartBeat()
			if nerr != nil {
				cc.ev.ProceedNicoEvent(&Event{
					Type:    EventTypeErr,
					Content: nerr,
				})
				continue
			}
			cc.ev.ProceedNicoEvent(&Event{
				Type:    EventTypeHeartBeatGot,
				Content: hbv,
			})
		}
	}
}

// FetchPostKey gets postkey using getpostkey API
func (cc *CommentConnection) FetchPostKey() Error {
	if cc.lv.Account == nil {
		return MakeError(ErrOther, "nil account in LiveWaku")
	}
	if cc.lv.BroadID == "" {
		return MakeError(ErrOther, "BroadID is not set")
	}

	c, nicoerr := NewNicoClient(cc.lv.Account)
	if nicoerr != nil {
		return nicoerr
	}

	url := fmt.Sprintf(
		"http://live.nicovideo.jp/api/getpostkey?thread=%s&block_no=%d",
		cc.lv.CommentServer.Thread, cc.block)
	res, err := c.Get(url)
	if err != nil {
		return ErrFromStdErr(err)
	}
	defer res.Body.Close()

	allb, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ErrFromStdErr(err)
	}

	pk := string(allb[8:])
	if pk == "" {
		return MakeError(ErrOther, "failed to get postkey")
	}

	cc.lv.PostKey = pk

	return nil
}

// SendComment sends comment to current comment connection
func (cc *CommentConnection) SendComment(text string, iyayo bool) Error {
	if cc.lv.PostKey == "" {
		return MakeError(ErrOther, "no postkey in livewaku")
	}
	if text == "" {
		return MakeError(ErrOther, "empty text")
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
	defer cc.wmu.Unlock()

	fmt.Fprint(cc.rw, sdcomm)
	err := cc.rw.Flush()
	if err != nil {
		return ErrFromStdErr(err)
	}

	return nil
}

// Disconnect close and disconnect
// terminate all goroutines and wait to exit
func (cc *CommentConnection) Disconnect() Error {
	if cc.disconnecting {
		return MakeError(ErrOther, "already disconnecting")
	}
	cc.disconnecting = true
	defer func() { cc.disconnecting = false }()

	cc.postKeyTmr.Stop()
	cc.heartbeatTmr.Stop()

	cc.cancel()
	cc.conn.Close()

	cc.wg.Wait()

	cc.ev.ProceedNicoEvent(&Event{
		Type:    EventTypeClose,
		Content: nil,
	})

	return nil
}
