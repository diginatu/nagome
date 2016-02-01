package nicolive

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	accountFileName = "userData.yml"
)

// Account is a niconico account
type Account struct {
	Mail        string
	Pass        string
	Usersession string
}

func (a *Account) String() string {
	i, l := 5, len(a.Mail)
	if i > l {
		i = l
	}
	return fmt.Sprintf("Account(%s..)", a.Mail[0:i])
}

// SaveAccount save Account to a file
func (a *Account) SaveAccount(filePath string) error {
	d, err := yaml.Marshal(a)
	if err != nil {
		return err
	}
	fmt.Printf("dump:\n%s\n\n", string(d))

	err = ioutil.WriteFile(filePath, d, 0600)
	if err != nil {
		return err
	}

	return nil
}

// LoadAccount lead from a file and returns a pointer to Account
func (a *Account) LoadAccount(filePath string) error {
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
