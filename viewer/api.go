package viewer

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/diginatu/nagome/nicolive"
)

// Message is base API struct for plugin
type Message struct {
	Domain  string          `json:"domain"`
	Command string          `json:"command"`
	Content json.RawMessage `json:"content,omitempty"` // The structure of Content is depend on the Command (and Domain).

	plgno int
}

func (m *Message) String() string {
	return fmt.Sprintf("{%s %s plug:%d}", m.Domain, m.Command, m.plgno)
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
		plgno:   -1,
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
	CommNagomeUserUpdate   = "User.Update" // CommNagomeUserUpdate is Emitted when User info is updated by fetching or setting name etc.

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

	CommQuerySettingsSetCurrent = "Settings.SetCurrent" // Set settings to current slot.
	CommQuerySettingsSetAll     = "Settings.SetAll"     // Set all slots of settings.

	CommQueryPlugEnable = "Plug.Enable" // Enable or disable a plugin.

	CommQueryUserSet     = "User.Set"     // Set user info like name to the DB.
	CommQueryUserSetName = "User.SetName" // Set user name to the DB.
	CommQueryUserDelete  = "User.Delete"  // Delete user info from the DB.
	CommQueryUserFetch   = "User.Fetch"   // Fetch user name from web page and update the internal user database.

	// DomainUI
	// Event to be processed by UI plugin.
	CommUINotification  = "Notification"
	CommUIClearComments = "ClearComments"
	CommUIConfigAccount = "ConfigAccount" // Open the window of account setting or suggest user to configure it.

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

	CommDirectUserGet = "User.Get" // Get user info from the user DB.

	// Nagome to plugin
	CommDirectngmPlugEnabled  = "Plug.Enabled"  // Sent when the plugin is enabled.
	CommDirectngmPlugDisabled = "Plug.Disabled" // Sent when the plugin is disabled.
	CommDirectngmPlugList     = "Plug.List"

	CommDirectngmSettingsCurrent = "Settings.Current"
	CommDirectngmSettingsAll     = "Settings.All"

	CommDirectngmUserGet = "User.Get"
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

// CtNagomeUserUpdate is a content of CommNagomeUserUpdate
type CtNagomeUserUpdate nicolive.User

// CtQueryBroadConnect is a content of CommQueryBroadConnect
type CtQueryBroadConnect struct {
	BroadID string `json:"broad_id"`
	RetryN  int    `json:"retry_n,omitempty"`
}

// type of CtQueryBroadSendComment
const (
	CtQueryBroadSendCommentTypeGeneral string = "General"
	CtQueryBroadSendCommentTypeOwner          = "Owner" // ignored if the user is not the owner
)

// CtQueryBroadSendComment is a content of CommQueryBroadSendComment
type CtQueryBroadSendComment struct {
	Text  string `json:"text"`
	Iyayo bool   `json:"iyayo"`
	Type  string `json:"type,omitempty"` // if omitted, automatically selected depend on the settings
}

// CtQueryAccountSet is a content of CommQueryAccountSet
type CtQueryAccountSet nicolive.Account

// CtQueryLogPrint is a content of CommQueryLogPrint
type CtQueryLogPrint struct {
	Text string `json:"text"`
}

// CtQuerySettingsSetCurrent is a content of CommQuerySettingsSetCurrent
type CtQuerySettingsSetCurrent SettingsSlot

// CtQuerySettingsSetAll is a content of CommQuerySettingsSetAll
type CtQuerySettingsSetAll SettingsSlots

// CtQueryPlugEnable is a content of CommQueryPlugEnable
type CtQueryPlugEnable struct {
	No     int  `json:"no"`
	Enable bool `json:"enable"`
}

// CtQueryUserSet is a content for CommQueryUserSet
type CtQueryUserSet nicolive.User

// CtQueryUserSetName is a content for CommQueryUserSetName
type CtQueryUserSetName struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CtQueryUserDelete is a content for CommQueryUserDelete
type CtQueryUserDelete struct {
	ID string `json:"id"`
}

// CtQueryUserFetch is a content for CommQueryUserFetch
type CtQueryUserFetch struct {
	ID string `json:"id"`
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

// CtUINotification is a content of CommUINotification
type CtUINotification struct {
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

// type of CtUINotification
const (
	CtUINotificationTypeInfo string = "Info"
	CtUINotificationTypeWarn        = "Warn"
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

// CtDirectUserGet is a content for CommDirectUserGet
type CtDirectUserGet struct {
	ID string `json:"id"`
}

// CtDirectngmUserGet is a content for CommDirectngmUserGet
type CtDirectngmUserGet nicolive.User
