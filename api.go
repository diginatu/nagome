package main

import (
	"encoding/json"
	"time"

	"github.com/diginatu/nagome/nicolive"
)

// Message is base API struct for plugin
type Message struct {
	// Domain that includes following parameters
	Domain string `json:"domain"`
	// Command
	Command string `json:"command"`
	// Elements type of Content is depend on witch Command is used
	Content json.RawMessage `json:"content,omitempty"`

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
	DomainDirect         = "nagome_direct" // DomainDirect is special domain.

	// Adding DomainFilterSuffix to the end of domain name in "depends" in your plugin.yml enables filtering messages by the plugin.
	// If there is a plugin that depends on filtering domain, Nagome will send messages that is in the domain to only this plugin.
	// This can used for even NagomeQuery but
	DomainFilterSuffix = "@filter"
)

// Command names
const (
	// DomainNagome
	CommNagomeOpen      = "Nagome.Open"
	CommNagomeClose     = "Nagome.Close"
	CommNagomeBroadInfo = "Nagome.BroadInfo"
	CommNagomeSend      = "Nagome.Send"

	// DomainComment
	// This domain is for only sending comments
	CommCommentGot = "Comment.Got"

	// DomainQuery
	// Query from plugin to Nagome
	CommQueryBroadConnect     = "Broad.Connect"
	CommQueryBroadDisconnect  = "Broad.Disconnect"
	CommQueryBroadSendComment = "Broad.SendComment"

	CommQueryAccountSet   = "Account.Set"   // set the given content value as account values.
	CommQueryAccountLogin = "Account.Login" // login and set the user session to account.
	CommQueryAccountLoad  = "Account.Load"
	CommQueryAccountSave  = "Account.Save"

	CommQueryLogPrint = "Log.Print" // print string using logger of Nagome

	// DomainUI
	// Event to be processed by UI plugin
	CommUIDialog string = "UI.Dialog"

	// DomainDirect
	// It's messages is sent between a plugin and Nagome.  Can not be filtered.
	CommDirectEnabled  = "Direct.Enabled"  // sent when the plugin is enabled.
	CommDirectDisabled = "Direct.Disabled" // sent when the plugin is disabled.
	CommDirectNo       = "Direct.No"       // tell plugin number to Nagome when the connection started.  (TCP at first time only)
)

// Contents
//
// Contents in the Message API

// CtNagomeBroadInfo is a content of CommNagomeBroadInfo
type CtNagomeBroadInfo nicolive.HeartbeatValue

// CtQueryBroadConnect is a content of CommQueryBroadConnect
type CtQueryBroadConnect struct {
	BroadID string
}

// CtQueryBroadSendComment is a content of CommQueryBroadSendComment
type CtQueryBroadSendComment struct {
	Text  string
	Iyayo bool
}

// CtQueryAccountSet is a content of CommQueryAccountSet
type CtQueryAccountSet nicolive.Account

// CtQueryLogPrint is a content of CommQueryLogPrint
type CtQueryLogPrint struct {
	Text string
}

// A CtCommentGot is a content of CommCommentGot
type CtCommentGot struct {
	No            int
	Date          time.Time
	UserID        string
	UserName      string
	Raw           string // raw comment text
	Comment       string // html format
	IsPremium     bool
	IsBroadcaster bool
	IsStaff       bool
	IsAnonymity   bool
	Score         int
}

// CtUIDialog is a content of CommUIDialog
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

// CtDirectNo is a content for CommDirectNo
type CtDirectNo struct {
	No int
}
