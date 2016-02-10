package nicolive

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

// UserSessionLoginClient manages UserSessionLogin.
type UserSessionLoginClient struct {
	mail   string
	pass   string
	client http.Client
}

// NewUserSessionLoginClient makes new userSessionLoginClient and returns pointer to it.
func NewUserSessionLoginClient(mail, pass string) *UserSessionLoginClient {
	jar, _ := cookiejar.New(nil)
	return &UserSessionLoginClient{
		mail:   mail,
		pass:   pass,
		client: http.Client{Jar: jar},
	}
}

// Request login to the niconico and set UserSession
func (cl *UserSessionLoginClient) Request() (string, error) {
	if cl.mail == "" || cl.pass == "" {
		return "", fmt.Errorf("mail or pass is not set")
	}

	params := url.Values{
		"mail":     []string{cl.mail},
		"password": []string{cl.pass},
	}
	resp, err := cl.client.PostForm("https://secure.nicovideo.jp/secure/login?site=nicolive", params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	nicoURL, err := url.Parse("http://nicovideo.jp")
	if err != nil {
		return "", err
	}
	for _, ck := range cl.client.Jar.Cookies(nicoURL) {
		if ck.Name == "user_session" {
			if ck.Value != "deleted" && ck.Value != "" {
				return ck.Value, nil
			}
		}
	}
	return "", fmt.Errorf("failed log in to niconico")
}
