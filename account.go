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

func saveAccount(account *Account) {
	d, err := yaml.Marshal(account)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	fmt.Printf("dump:\n%s\n\n", string(d))

	err = ioutil.WriteFile(filepath.Join(App.SavePath, "userData.yml"), d, 0600)
	if err != nil {
		panic(err)
	}
}

func loadAccount() *Account {
	a := new(Account)
	d, err := ioutil.ReadFile(filepath.Join(App.SavePath, "userData.yml"))
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(d, a)
	if err != nil {
		fmt.Printf("error: %v", err)
	}

	return a
}
