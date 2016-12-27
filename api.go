package main

import (
	"encoding/json"
	"log"
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
	var conj json.RawMessage
	var err error

	if con == nil {
		conj = nil
	} else {
		conj, err = json.Marshal(con)
		if err != nil {
			return nil, err
		}
	}

	m := &Message{
		Domain:  dom,
		Command: com,
		Content: conj,
		prgno:   -1,
	}
	return m, nil
}

// NewMessageMust is same as NewMessage but assume no error.
func NewMessageMust(dom, com string, con interface{}) *Message {
	m, err := NewMessage(dom, com, con)
	if err != nil {
		log.Fatalln(err)
	}
	return m
}

// Dimain names
const (
	DomainNagome    = "nagome"
	DomainQuery     = "nagome_query"
	DomainComment   = "nagome_comment"
	DomainUI        = "nagome_ui"
	DomainAntenna   = "nagome_antenna"
	DomainDirect    = "nagome_direct"    // DomainDirect is a special domain (from plugin).
	DomainDirectngm = "nagome_directngm" // DomainDirectNgm is a domain for direct message from Nagome.

	// Adding DomainSuffixFilter to the end of domain name in "subscribe" in your plugin.yml enables filtering messages by the plugin.
	DomainSuffixFilter = "@filter"
)

// Command names
const (
	// DomainNagome
	CommNagomeBroadOpen    = "Broad.Open"
	CommNagomeBroadClose   = "Broad.Close"
	CommNagomeBroadInfo    = "Broad.Info"
	CommNagomeCommentSend  = "Comment.Send"
	CommNagomeAntennaOpen  = "Antenna.Open"
	CommNagomeAntennaClose = "Antenna.Close"

	// DomainComment
	// This domain is for only sending comments.
	CommCommentGot = "Got"

	// DomainQuery
	// Query from plugin to Nagome.
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

	CommQueryPlugEnable = "Plug.Enable" // Enable or disable a plugin.

	// DomainUI
	// Event to be processed by UI plugin.
	CommUIDialog        = "Dialog"
	CommUIClearComments = "ClearComments"

	// DomainAntenna
	// All antenna items (started live).
	CommAntennaGot = "Got"

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

// CtNagomeBroadOpen is a content of CommNagomeBroadOpen
type CtNagomeBroadOpen struct {
	BroadID     string `json:"broad_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CommunityID string `json:"community_id"`
	OwnerID     string `json:"owner_id"`
	OwnerName   string `json:"owner_name"`
	OwnerBroad  bool   `json:"owner_broad"`

	OpenTime  time.Time `json:"open_time"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

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

// CtQueryPlugEnable is a content of CommQueryPlugEnable
type CtQueryPlugEnable struct {
	No     int  `json:"no"`
	Enable bool `json:"enable"`
}

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

// CtAntennaGot is a content of CommAntennaGot
type CtAntennaGot struct {
	BroadID     string `json:"broad_id"`
	CommunityID string `json:"community_id"`
	UserID      string `json:"user_id"`
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
	Plugins *[]*Plugin `json:"plugins"`
}

// CtDirectngmSettingsCurrent is a content for CommDirectngmSettingsCurrent
type CtDirectngmSettingsCurrent SettingsSlot

// CtDirectngmSettingsAll is a content for CommDirectngmSettingsAll
type CtDirectngmSettingsAll SettingsSlots
