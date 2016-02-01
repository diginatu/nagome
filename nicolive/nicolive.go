// Package nicolive provides access way to NicoNama API,
// corresponding structure and other features.
package nicolive

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"gopkg.in/xmlpath.v1"
)

var (
	statusPath    = xmlpath.MustCompile("//@status")
	errorCodePath = xmlpath.MustCompile("//error/code")
)

// NewNicoClient makes new http.Client with usersession
func NewNicoClient(a *Account) (*http.Client, error) {
	nicoURL, err := url.Parse("http://nicovideo.jp")
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
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
