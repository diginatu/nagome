// Package nicolive provides access way to NicoNama API,
// corresponding structures and other features.
package nicolive

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"runtime"

	"gopkg.in/xmlpath.v2"
)

var (
	statusXMLPath    = xmlpath.MustCompile("//@status")
	errorCodeXMLPath = xmlpath.MustCompile("//error/code")
	errorDescXMLPath = xmlpath.MustCompile("//error/description")
)

func init() {
}

// NewNicoClient makes new http.Client with usersession
func NewNicoClient(a *Account) (*http.Client, Error) {
	if a.Usersession == "" {
		return nil, MakeError(ErrOther, "no usersession")
	}

	nicoURL, err := url.Parse("http://nicovideo.jp")
	if err != nil {
		return nil, ErrFromStdErr(err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, ErrFromStdErr(err)
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

func caller(sk int) string {
	_, file, line, ok := runtime.Caller(sk)
	short := file
	if !ok {
		short = "???"
		line = 0
	} else {
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
	}
	return fmt.Sprintf("%s:%d", short, line)
}
