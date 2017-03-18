package nicolive

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"gopkg.in/yaml.v2"
)

const (
	loginAddr    = "https://secure.nicovideo.jp/secure/login?site=nicolive"
	nicoBaseAddr = "http://nicovideo.jp"
)

var (
	nicoBaseURL *url.URL
)

func init() {
	var err error
	nicoBaseURL, err = url.Parse(nicoBaseAddr)
	if err != nil {
		panic(err)
	}
}

// Account is a niconico account
type Account struct {
	Mail        string `yaml:"mail"`
	Pass        string `yaml:"pass"`
	Usersession string `yaml:"usersession"`
	client      *http.Client
}

// NewAccount makes new account with a http client.
func NewAccount(mail, pass, usersession string) *Account {
	a := &Account{mail, pass, usersession, nil}
	a.UpdateClient()
	return a
}

// UpdateClient updates Client with its Usersession.
// If the Usersession of the account is empty string, clear the cookies jar.
func (a *Account) UpdateClient() {
	if a.Usersession == "" {
		if a.client != nil {
			a.client.Jar = nil
		}
		return
	}

	if a.client == nil {
		a.client = &http.Client{}
	}
	if a.client.Jar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			panic(err) // I think an error is never occurred.
		}
		a.client.Jar = jar
	}

	a.client.Jar.SetCookies(nicoBaseURL, []*http.Cookie{
		{
			Domain: nicoBaseURL.Host,
			Path:   "/",
			Name:   "user_session",
			Value:  a.Usersession,
			Secure: false,
		},
	})
	return
}

func (a *Account) String() string {
	i, l := 5, len(a.Mail)
	if i > l {
		i = l
	}
	return fmt.Sprintf("Account{%s..}", a.Mail[0:i])
}

// Save save Account to a file
func (a *Account) Save(filePath string) error {
	d, err := yaml.Marshal(a)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, d, 0600)
	if err != nil {
		return err
	}

	return nil
}

// Load reads from a file and sets values
func (a *Account) Load(filePath string) error {
	d, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(d, a)
	if err != nil {
		return err
	}

	a.UpdateClient()
	return nil
}

// Login log in to niconico and update Usersession
func (a *Account) Login() error {
	return a.LoginImpl(loginAddr)
}

// LoginImpl is implementation of Login.
func (a *Account) LoginImpl(addr string) (err error) {
	if a.Mail == "" || a.Pass == "" {
		return MakeError(ErrOther, "invalid account : mail or pass is not set")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return ErrFromStdErr(err)
	}
	cl := http.Client{Jar: jar}

	params := url.Values{
		"mail":     []string{a.Mail},
		"password": []string{a.Pass},
	}
	resp, err := cl.PostForm(addr, params)
	if err != nil {
		return ErrFromStdErr(err)
	}
	defer func() {
		lerr := resp.Body.Close()
		if lerr != nil && err == nil {
			err = lerr
		}
	}()

	for _, ck := range cl.Jar.Cookies(nicoBaseURL) {
		if ck.Name == "user_session" {
			if ck.Value != "deleted" && ck.Value != "" {
				a.Usersession = ck.Value
				a.UpdateClient()
				return nil
			}
		}
	}

	return MakeError(ErrOther, "failed log in to niconico.  Could not find the key of the usersession cookie")
}
