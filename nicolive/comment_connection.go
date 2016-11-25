package nicolive

import (
	"context"
	"fmt"
	"html"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"

	"gopkg.in/xmlpath.v2"
)

const (
	postKeyDuration   = 10 * time.Second
	heartbeatDuration = 90 * time.Second
)

// A Comment is a received comment.
// It's send as a content of an EventTypeGot event.
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
	*connection

	lv     *LiveWaku
	ticket string
	svrDu  time.Duration
	block  int

	postKeyTmr   *time.Timer
	heartbeatTmr *time.Timer
	ConnectedTm  time.Time // The time when this connection started
}

// CommentConnect connects to nicolive and start receiving comment.
// liveWaku should have connection information which is able to get by fetchInformation()
func CommentConnect(ctx context.Context, lv *LiveWaku, ev EventReceiver) (*CommentConnection, error) {
	if lv.Account == nil {
		return nil, MakeError(ErrOther, "nil account in LiveWaku")
	}
	if lv.BroadID == "" {
		return nil, MakeError(ErrOther, "BroadID is not set")
	}

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
	}
	cc.connection = newConnection(
		net.JoinHostPort(lv.CommentServer.Addr, lv.CommentServer.Port),
		cc.proceedMessage, ev)

	nerr := cc.connection.Connect(ctx)
	if nerr != nil {
		return nil, nerr
	}

	cc.Wg.Add(1)
	go cc.timer()

	nerr = cc.connection.Send(fmt.Sprintf(
		"<thread thread=\"%s\" res_from=\"-1000\" version=\"20061206\" />\x00",
		cc.lv.CommentServer.Thread))
	if nerr != nil {
		go func() {
			_ = cc.Disconnect()
		}()
		return nil, nerr
	}

	return cc, nil
}

func (cc *CommentConnection) proceedMessage(m string) {
	commxmlr := strings.NewReader(m)
	rt, err := xmlpath.Parse(commxmlr)
	if err != nil {
		cc.Ev.ProceedNicoEvent(&Event{
			Type:    EventTypeErr,
			Content: err,
		})
		return
	}

	if strings.HasPrefix(m, "<thread ") {
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

		cc.Ev.ProceedNicoEvent(&Event{
			Type:    EventTypeOpen,
			Content: cc.lv,
		})

		return
	}
	if strings.HasPrefix(m, "<chat_result ") {
		if v, ok := xmlpath.MustCompile("/chat_result/@status").String(rt); ok {
			if v != "0" {
				cc.Ev.ProceedNicoEvent(&Event{
					Type:    EventTypeErr,
					Content: MakeError(ErrSendComment, "comment send error. chat_result status : "+v),
				})
				return
			}
			cc.Ev.ProceedNicoEvent(&Event{
				Type:    EventTypeSend,
				Content: nil,
			})
		}
		return
	}
	if strings.HasPrefix(m, "<chat ") {
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

		cc.Ev.ProceedNicoEvent(&Event{
			Type:    EventTypeGot,
			Content: comment,
		})

		if comment.IsCommand && comment.Comment == "/disconnect" {
			go func() {
				_ = cc.Disconnect()
			}()
			cc.Ev.ProceedNicoEvent(&Event{
				Type:    EventTypeWakuEnd,
				Content: *cc.lv,
			})
			return
		}
		return
	}

	cc.Ev.ProceedNicoEvent(&Event{
		Type:    EventTypeErr,
		Content: MakeError(ErrSendComment, "unknown stream : "+m),
	})
}

func (cc *CommentConnection) timer() {
	defer cc.Wg.Done()
	for {
		select {
		case <-cc.Ctx.Done():
			return
		case <-cc.postKeyTmr.C:
			cc.postKeyTmr.Reset(postKeyDuration)
			nerr := cc.FetchPostKey()
			if nerr != nil {
				cc.Ev.ProceedNicoEvent(&Event{
					Type:    EventTypeErr,
					Content: nerr,
				})
				continue
			}
		case <-cc.heartbeatTmr.C:
			cc.heartbeatTmr.Reset(heartbeatDuration)
			hbv, nerr := cc.lv.FetchHeartBeat()
			if nerr != nil {
				cc.Ev.ProceedNicoEvent(&Event{
					Type:    EventTypeErr,
					Content: nerr,
				})
				continue
			}
			cc.Ev.ProceedNicoEvent(&Event{
				Type:    EventTypeHeartBeatGot,
				Content: &hbv,
			})
		}
	}
}

// FetchPostKey gets postkey using getpostkey API
func (cc *CommentConnection) FetchPostKey() (err error) {
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
	defer func() {
		lerr := res.Body.Close()
		if lerr != nil && err == nil {
			err = lerr
		}
	}()

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
func (cc *CommentConnection) SendComment(text string, iyayo bool) error {
	if cc.lv.PostKey == "" {
		return MakeError(ErrOther, "no postkey in livewaku")
	}
	if text == "" {
		return MakeError(ErrOther, "empty text")
	}

	vpos := 100 * (time.Now().Add(cc.svrDu).Unix() - cc.lv.Stream.OpenTime.Unix())

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

	err := cc.connection.Send(sdcomm)
	if err != nil {
		return ErrFromStdErr(err)
	}

	return nil
}

// Disconnect quit all routines and disconnect.
func (cc *CommentConnection) Disconnect() error {
	cc.postKeyTmr.Stop()
	cc.heartbeatTmr.Stop()

	err := cc.connection.Disconnect()
	if err != nil {
		return err
	}

	cc.Ev.ProceedNicoEvent(&Event{
		Type:    EventTypeClose,
		Content: nil,
	})

	return nil
}
