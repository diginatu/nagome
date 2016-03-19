// Package nicolive provides access way to NicoNama API,
// corresponding structure and other features.
package nicolive

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"gopkg.in/xmlpath.v2"
)

var (
	statusXMLPath    = xmlpath.MustCompile("//@status")
	errorCodeXMLPath = xmlpath.MustCompile("//error/code")

	// Logger is used in nicolive to output logs
	Logger *log.Logger
	// EvReceiver is EventReceiver used in nicolive
	EvReceiver EventReceiver
)

func init() {
	Logger = log.New(os.Stderr, "", log.Lshortfile|log.Ltime)
	EvReceiver = &defaultEventReceiver{}
}

// Event is an event
type Event struct {
	EventString string
	Content     interface{}
}

func (e *Event) String() string {
	return e.EventString
}

// EventReceiver receive events and proceed
type EventReceiver interface {
	Proceed(*Event)
}

type defaultEventReceiver struct{}

func (der *defaultEventReceiver) Proceed(ev *Event) {
	Logger.Println(ev.EventString, ev.Content)
}

// NewNicoClient makes new http.Client with usersession
func NewNicoClient(a *Account) (*http.Client, NicoError) {
	if a.Usersession == "" {
		return nil, NicoErr(NicoErrOther, "no usersession", "usersession is empty in this accout")
	}

	nicoURL, err := url.Parse("http://nicovideo.jp")
	if err != nil {
		return nil, NicoErrFromStdErr(err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, NicoErrFromStdErr(err)
	}
	c := http.Client{Jar: jar}
	c.Jar.SetCookies(nicoURL, []*http.Cookie{
		&http.Cookie{
			Domain: nicoURL.Host,
			Path:   "/",
			Name:   "user_session",
			Value:  a.Usersession,
			Secure: false,
		},
	})
	return &c, nil
}
