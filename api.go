package main

import (
	"encoding/json"
	"time"

	"github.com/diginatu/nagome/nicolive"
)

// Message is base API struct for plugin
type Message struct {
	// Domain that includes following parameters
	Domain string
	// Command
	Command string
	// Elements type of Content is depend on witch Command is used
	Content json.RawMessage

	prgno int
}

// NewMessage returns new Message with the given values.
func NewMessage(dom, com string, con interface{}) (*Message, error) {
	conj, err := json.Marshal(con)
	if err != nil {
		return nil, err
	}

	m := &Message{
		Domain:  dom,
		Command: com,
		Content: conj,
	}
	return m, nil
}

// Dimain names
const (
	DomainNagome  string = "nagome"
	DomainQuery          = "nagome_query"
	DomainComment        = "nagome_comment"
	DomainUI             = "nagome_ui"

	FilterSuffix = "@filter"
)

// Command names
const (
	// DomainNagome
	CommNagomeOpen      = "Nagome.Open"
	CommNagomeClose     = "Nagome.Close"
	CommNagomeBroadInfo = "Nagome.BroadInfo"
	CommNagomeSend      = "Nagome.Send"

	// DomainComment
	CommCommentGot = "Comment.Got"

	// DomainQuery
	CommQueryBroadConnect     = "Broad.Connect"
	CommQueryBroadDisconnect  = "Broad.Disconnect"
	CommQueryBroadSendComment = "Broad.SendComment"

	CommQueryAccountSet   = "Account.Set" // set the given content value as account values.
	CommQueryAccountLogin = "Account.Login"
	CommQueryAccountLoad  = "Account.Load"
	CommQueryAccountSave  = "Account.Save"

	// DomainUI
	CommUIDialog string = "UI.Dialog"
)

// Contents
//
// Contents in the Message API

// CtQueryPluginNo is content of QueryPluginNo
// only for TCP at first time
type CtQueryPluginNo struct {
	No int
}

// CtNagomeBroadInfo is content of NagomeBroadInfo
type CtNagomeBroadInfo nicolive.HeartbeatValue

// CtQueryBroadConnect is content of QueryBroadConnect
type CtQueryBroadConnect struct {
	BroadID string
}

// CtQueryBroadSendComment is content of QueryBroadSendComment
type CtQueryBroadSendComment struct {
	Text  string
	Iyayo bool
}

// CtQueryAccountSet is content of QueryAccountSet
type CtQueryAccountSet nicolive.Account

// A CtCommentGot is a content of got comment
type CtCommentGot struct {
	No            int
	Date          time.Time
	UserID        string
	UserName      string
	Comment       string // html format
	IsPremium     bool
	IsBroadcaster bool
	IsStaff       bool
	IsAnonymity   bool
	Score         int
}

// CtUIDialog is content of dialog that nagome ask to open
type CtUIDialog struct {
	Type        string // select from below const string
	Title       string
	Description string
}

// type of CtUIDialog
const (
	CtUIDialogTypeInfo string = "Info"
	CtUIDialogTypeWarn        = "Warn"
)
