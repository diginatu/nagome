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
	DomainNagome    string = "nagome"
	DomainQuery            = "nagome_query"
	DomainComment          = "nagome_comment"
	DomainUI               = "nagome_ui"
	DomainDirect           = "nagome_direct"    // DomainDirect is a special domain (from plugin).
	DomainDirectngm        = "nagome_directngm" // DomainDirectNgm is a domain for direct message from Nagome.

	// Adding DomainFilterSuffix to the end of domain name in "depends" in your plugin.yml enables filtering messages by the plugin.
	// If there is a plugin that depends on filtering domain, Nagome will send a message of the domain to only the filtering plugin.
	DomainFilterSuffix = "@filter"
)

// Command names
const (
	// DomainNagome
	CommNagomeBroadOpen   = "Broad.Open"
	CommNagomeBroadClose  = "Broad.Close"
	CommNagomeBroadInfo   = "Broad.Info"
	CommNagomeCommentSend = "Comment.Send"

	// DomainComment
	// This domain is for only sending comments
	CommCommentGot = "Got"

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

	CommQuerySettingsSet    = "Settings.Set"    // Set settings to current slot.
	CommQuerySettingsSetAll = "Settings.SetAll" // Set all slots of settings.

	// DomainUI
	// Event to be processed by UI plugin
	CommUIDialog        = "Dialog"
	CommUIClearComments = "ClearComments"

	// DomainDirect (special domain)
	// The messages is sent between a plugin and Nagome.  It is not broadcasted and can not be filtered.

	// plugin to Nagome
	CommDirectNo       = "No"        // Tell plugin number to Nagome when the connection started.  (TCP at first time only)
	CommDirectPlugList = "Plug.List" // Request a list of plugins.

	CommDirectSettingsCurrent = "Settings.Current" // Request current settings message.
	CommDirectSettingsAll     = "Settings.All"     // Request all slots of settings message.

	// Nagome to plugin
	CommDirectngmPlugEnabled  = "Plug.Enabled"  // Sent when the plugin is enabled.
	CommDirectngmPlugDisabled = "Plug.Disabled" // Sent when the plugin is disabled.
	CommDirectngmPlugList     = "Plug.List"

	CommDirectngmSettingsCurrent = "Settings.Current"
	CommDirectngmSettingsAll     = "Settings.All"
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

// CtQuerySettingsSet is a content of CommQuerySettingsSet
type CtQuerySettingsSet SettingsSlot

// CtQuerySettingsSetSlots is a content of CommQuerySettingsSetSlots
type CtQuerySettingsSetSlots SettingsSlots

// A CtCommentGot is a content of CommCommentGot
type CtCommentGot struct {
	No      int       `json:"no"`
	Date    time.Time `json:"date"`
	Raw     string    `json:"raw"`
	Comment string    `json:"comment"`

	UserID           string `json:"user_id"`
	UserName         string `json:"user_name"`
	UserThumbnailURL string `json:"user_thumbnail_url,omitempty"`
	Score            int    `json:"score,omitempty"`
	IsPremium        bool   `json:"is_premium"`
	IsBroadcaster    bool   `json:"is_broadcaster"`
	IsStaff          bool   `json:"is_staff"`
	IsAnonymity      bool   `json:"is_anonymity"`
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

// CtDirectngmPlugList is a content for CommDirectngmPlugList
type CtDirectngmPlugList struct {
	Plugins *[]*plugin `json:"plugins"`
}

// CtDirectngmSettingsCurrent is a content for CommDirectSettingsCurrent
type CtDirectngmSettingsCurrent SettingsSlot

// CtDirectngmSettingsAll is a content for CommDirectSettingsSlots
type CtDirectngmSettingsAll SettingsSlots
