package main

import (
	"encoding/json"
	"time"

	"github.com/diginatu/nagome/nicolive"
)

// Message is base API struct for plugin
type Message struct {
	Domain  string          `json:"domain"`
	Command string          `json:"command"`
	Content json.RawMessage `json:"content,omitempty"` // The structure of Content is depend on the Command (and Domain).

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

	CommQueryAccountSet   = "Account.Set"   // Set the given content value as account values.
	CommQueryAccountLogin = "Account.Login" // Login and set the user session to account.
	CommQueryAccountLoad  = "Account.Load"
	CommQueryAccountSave  = "Account.Save"

	CommQueryLogPrint = "Log.Print" // Print string using logger of Nagome

	// DomainUI
	// Event to be processed by UI plugin
	CommUIDialog = "UI.Dialog"

	// DomainDirect (special domain)
	// Its messages is sent between a plugin and Nagome.  It is not broadcasted and can not be filtered.

	// plugin to Nagome
	CommDirectNo          = "Direct.No"          // Tell plugin number to Nagome when the connection started.  (TCP at first time only)
	CommDirectReqListPlug = "Direct.ReqListPlug" // Request a list of plugins.

	// Nagome to plugin
	CommDirectEnabled  = "Direct.Enabled"  // Sent when the plugin is enabled.
	CommDirectDisabled = "Direct.Disabled" // Sent when the plugin is disabled.
	CommDirectListPlug = "Direct.ListPlug"
)

// Contents
//
// Contents in the Message API

// CtNagomeBroadInfo is a content of CommNagomeBroadInfo
type CtNagomeBroadInfo struct {
	WatchCount   string `json:"watch_count"`
	CommentCount string `json:"comment_count"`
}

// CtQueryBroadConnect is a content of CommQueryBroadConnect
type CtQueryBroadConnect struct {
	BroadID string `json:"broad_id"`
}

// CtQueryBroadSendComment is a content of CommQueryBroadSendComment
type CtQueryBroadSendComment struct {
	Text  string `json:"text"`
	Iyayo bool   `json:"iyayo"`
}

// CtQueryAccountSet is a content of CommQueryAccountSet
type CtQueryAccountSet nicolive.Account

// CtQueryLogPrint is a content of CommQueryLogPrint
type CtQueryLogPrint struct {
	Text string `json:"text"`
}

// A CtCommentGot is a content of CommCommentGot
type CtCommentGot struct {
	No            int       `json:"no"`
	Date          time.Time `json:"date"`
	UserID        string    `json:"user_id"`
	UserName      string    `json:"user_name"`
	Raw           string    `json:"raw"`
	Comment       string    `json:"comment"`
	IsPremium     bool      `json:"is_premium"`
	IsBroadcaster bool      `json:"is_broadcaster"`
	IsStaff       bool      `json:"is_staff"`
	IsAnonymity   bool      `json:"is_anonymity"`
	Score         int       `json:"score"`
}

// CtUIDialog is a content of CommUIDialog
type CtUIDialog struct {
	// Select type from below const string
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// type of CtUIDialog
const (
	CtUIDialogTypeInfo string = "Info"
	CtUIDialogTypeWarn        = "Warn"
)

// CtDirectNo is a content for CommDirectNo
type CtDirectNo struct {
	No int `json:"no"`
}

// CtDirectListPlug is a content for CommDirectListPlug
type CtDirectListPlug struct {
	Plugins *[]*plugin `json:"plugins"`
}
