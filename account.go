package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Account is a niconico account
type Account struct {
	Mail        string
	Pass        string
	Usersession string
}

func (a *Account) String() string {
	return fmt.Sprintf("%#v", a)
}

func saveAccount(account *Account) error {
	d, err := yaml.Marshal(account)
	if err != nil {
		return err
	}
	fmt.Printf("dump:\n%s\n\n", string(d))

	err = ioutil.WriteFile(filepath.Join(App.SavePath, "userData.yml"), d, 0600)
	if err != nil {
		return err
	}

	return nil
}

func loadAccount() (*Account, error) {
	a := new(Account)
	d, err := ioutil.ReadFile(filepath.Join(App.SavePath, "userData.yml"))
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(d, a)
	if err != nil {
		return nil, err
	}

	return a, nil
}
