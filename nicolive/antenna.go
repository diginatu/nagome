package nicolive

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/xmlpath.v2"
)

// An AntennaItem is a started live broadcast.
// It's send as a content of an EventTypeAntennaGot event.
type AntennaItem struct {
	BroadID, CommunityID, UserID string
}

// Antenna manages starting broadcast antenna connection.
type Antenna struct {
	*connection

	ac                 *Account
	ticket             string
	addr, port, thread string

	MyCommunities []string
}

// ConnectAntenna connects to antenna and return new Antenna.
func ConnectAntenna(ctx context.Context, ac *Account, ev EventReceiver) (*Antenna, error) {
	a := NewAntenna(ac)
	var err error

	err = a.Login()
	if err != nil {
		return nil, err
	}
	err = a.Admin()
	if err != nil {
		return nil, err
	}
	err = a.Connect(ctx, ev)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// NewAntenna creates new Antenna.
func NewAntenna(ac *Account) *Antenna {
	return &Antenna{
		ac: ac,
	}
}

// Login logs in to the antenna connection.
func (a *Antenna) Login() error {
	cl := new(http.Client)
	vl := url.Values{"mail": {a.ac.Mail}, "password": {a.ac.Pass}}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://secure.nicovideo.jp/secure/login?site=nicolive_antenna",
		strings.NewReader(vl.Encode()),
	)
	if err != nil {
		return ErrFromStdErr(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := cl.Do(req)
	if err != nil {
		return ErrFromStdErr(err)
	}
	defer res.Body.Close()

	return a.loginParseProc(res.Body)
}
func (a *Antenna) loginParseProc(r io.Reader) error {
	root, err := xmlpath.Parse(r)
	if err != nil {
		return ErrFromStdErr(err)
	}

	if v, ok := xmlPathStatus.String(root); ok {
		if v != "ok" {
			if v, ok := xmlPathErrorCode.String(root); ok {
				desc, _ := xmlPathErrorDesc.String(root)
				return MakeError(ErrOther, v+": "+desc)
			}
			return MakeError(ErrOther, "request failed with unknown error")
		}
	}

	if v, ok := xmlpath.MustCompile("//ticket").String(root); ok {
		if v == "" {
			return MakeError(ErrIncorrectAccount, "incorrect account")
		}
		a.ticket = v
	}

	return nil
}

// Admin gets favorite communities and information to connect.
func (a *Antenna) Admin() error {
	if a.ticket == "" {
		return MakeError(ErrOther, "The ticket is not set.  Login first.")
	}

	cl := new(http.Client)
	vl := url.Values{"ticket": {a.ticket}}

	req, err := http.NewRequest(
		http.MethodPost,
		"http://live.nicovideo.jp/api/getalertstatus",
		strings.NewReader(vl.Encode()),
	)
	if err != nil {
		return ErrFromStdErr(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := cl.Do(req)
	if err != nil {
		return ErrFromStdErr(err)
	}
	defer res.Body.Close()

	return a.adminParseProc(res.Body)
}
func (a *Antenna) adminParseProc(r io.Reader) error {
	root, err := xmlpath.Parse(r)
	if err != nil {
		return ErrFromStdErr(err)
	}

	if v, ok := xmlPathStatus.String(root); ok {
		if v != "ok" {
			if v, ok := xmlPathErrorCode.String(root); ok {
				desc, _ := xmlPathErrorDesc.String(root)
				return MakeError(ErrOther, v+": "+desc)
			}
			return MakeError(ErrOther, "request failed with unknown error")
		}
	}

	if v, ok := xmlpath.MustCompile("//ms/addr").String(root); ok {
		a.addr = v
	}
	if v, ok := xmlpath.MustCompile("//ms/port").String(root); ok {
		a.port = v
	}
	if v, ok := xmlpath.MustCompile("//ms/thread").String(root); ok {
		a.thread = v
	}

	return nil
}

// Connect connects to antenna
func (a *Antenna) Connect(ctx context.Context, ev EventReceiver) error {
	if a.addr == "" || a.port == "" || a.thread == "" {
		return MakeError(ErrOther, "Connection info is not set.  Do Admin() first.")
	}

	if ev == nil {
		ev = &defaultEventReceiver{}
	}

	a.connection = newConnection(
		net.JoinHostPort(a.addr, a.port),
		a.proceedMessage, ev)

	var err error

	err = a.connection.Connect(ctx)
	if err != nil {
		a.conn = nil
		return err
	}

	err = a.Send(fmt.Sprintf(
		"<thread thread=\"%s\" res_from=\"-1\" version=\"20061206\" />\x00",
		a.thread))

	if err != nil {
		go a.Disconnect()
		return ErrFromStdErr(err)
	}

	return nil
}

func (a *Antenna) proceedMessage(m string) {
	xmlr := strings.NewReader(m)
	rt, err := xmlpath.Parse(xmlr)
	if err != nil {
		a.Ev.ProceedNicoEvent(&Event{
			Type:    EventTypeAntennaErr,
			Content: err,
		})
		return
	}

	if strings.HasPrefix(m, "<thread ") {
		a.Ev.ProceedNicoEvent(&Event{
			Type:    EventTypeAntennaOpen,
			Content: nil,
		})
		return
	}
	if strings.HasPrefix(m, "<chat ") {
		if v, ok := xmlpath.MustCompile("/chat").String(rt); ok {
			av := strings.Split(v, ",")
			ai := &AntennaItem{av[0], av[1], av[2]}
			a.Ev.ProceedNicoEvent(&Event{
				Type:    EventTypeAntennaGot,
				Content: ai,
			})
		}
		return
	}

	a.Ev.ProceedNicoEvent(&Event{
		Type:    EventTypeAntennaErr,
		Content: MakeError(ErrSendComment, "unknown stream : "+m),
	})
}

// Disconnect quit all routines and disconnect.
func (a *Antenna) Disconnect() error {
	if a.conn == nil {
		return MakeError(ErrOther, "Antenna is not connected.")
	}

	err := a.connection.Disconnect()
	if err != nil {
		return err
	}

	a.Ev.ProceedNicoEvent(&Event{
		Type:    EventTypeAntennaClose,
		Content: nil,
	})

	a.conn = nil
	return nil
}
