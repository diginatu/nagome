package nicolive

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"gopkg.in/yaml.v2"
)

// Account is a niconico account
type Account struct {
	Mail        string `yaml:"mail"`
	Pass        string `yaml:"pass"`
	Usersession string `yaml:"-"`
}

func (a Account) String() string {
	i, l := 5, len(a.Mail)
	if i > l {
		i = l
	}
	return fmt.Sprintf("Account{%s..}", a.Mail[0:i])
}

// Save save Account to a file
func (a Account) Save(filePath string) error {
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

// Load lead from a file and returns a pointer to Account
func (a *Account) Load(filePath string) error {
	d, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(d, a)
	if err != nil {
		return err
	}

	return nil
}

// Login log in to niconico and update Usersession
func (a *Account) Login() NicoError {
	if a.Mail == "" || a.Pass == "" {
		return NicoErr(NicoErrOther, "invalid account info", "mail or pass is not set")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return NicoErrFromStdErr(err)
	}
	cl := http.Client{Jar: jar}

	params := url.Values{
		"mail":     []string{a.Mail},
		"password": []string{a.Pass},
	}
	resp, err := cl.PostForm("https://secure.nicovideo.jp/secure/login?site=nicolive", params)
	if err != nil {
		return NicoErrFromStdErr(err)
	}
	defer resp.Body.Close()

	nicoURL, err := url.Parse("http://nicovideo.jp")
	if err != nil {
		return NicoErrFromStdErr(err)
	}
	for _, ck := range cl.Jar.Cookies(nicoURL) {
		if ck.Name == "user_session" {
			if ck.Value != "deleted" && ck.Value != "" {
				a.Usersession = ck.Value
				return nil
			}
		}
	}

	return NicoErr(NicoErrOther, "login error", "failed log in to niconico")
}
