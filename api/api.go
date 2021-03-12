package api

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Message is base API struct for plugin
type Message struct {
	Domain  string          `json:"domain,omitempty"`
	Command string          `json:"command,omitempty"`
	Content json.RawMessage `json:"content,omitempty"` // The structure of Content is depend on the Command (and Domain).

	Plgno int `json:"-"`
}

func (m *Message) String() string {
	return fmt.Sprintf("{%s %s plug:%d}", m.Domain, m.Command, m.Plgno)
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
		Plgno:   -1,
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
	DomainDirect    = "nagome_direct"    // DomainDirect is a special domain (from plugin).
	DomainDirectngm = "nagome_directngm" // DomainDirectNgm is a domain for direct message from Nagome.

	// Adding DomainSuffixFilter to the end of domain name in "subscribe" in your plugin.yml enables filtering messages by the plugin.
	DomainSuffixFilter = "@filter"
)

// Command names
const (
	// DomainNagome
	// Event that is mainly sent from Nagome
	CommNagomeBroadOpen   = "Broad.Open"
	CommNagomeBroadClose  = "Broad.Close"
	CommNagomeBroadInfo   = "Broad.Info"
	CommNagomeCommentSend = "Comment.Send"
	CommNagomeUserUpdate  = "User.Update" // CommNagomeUserUpdate is Emitted when User info is updated by fetching or setting name etc.

	// DomainComment
	// Event that is mainly sent from Nagome
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

	// DomainDirect (special domain)
	// The messages is sent between a plugin and Nagome.  It is not broadcasted and can not be filtered.

	// from plugin to Nagome
	CommDirectAppVersion = "App.Version"

	CommDirectNo       = "No"        // Tell plugin number to Nagome when the connection started.  (TCP at first time only)
	CommDirectPlugList = "Plug.List" // Request a list of plugins.

	CommDirectSettingsCurrent = "Settings.Current" // Request current settings message.
	CommDirectSettingsAll     = "Settings.All"     // Request all slots of settings message.

	CommDirectUserGet = "User.Get" // Get user info from the user DB.

	// from Nagome to plugin
	CommDirectngmAppVersion = "App.Version"

	CommDirectngmPlugEnabled  = "Plug.Enabled"  // Sent when the plugin is enabled.
	CommDirectngmPlugDisabled = "Plug.Disabled" // Sent when the plugin is disabled.
	CommDirectngmPlugList     = "Plug.List"

	CommDirectngmSettingsCurrent = "Settings.Current"
	CommDirectngmSettingsAll     = "Settings.All"

	CommDirectngmUserGet = "User.Get"
)

// Live Platforms
const (
	PlatformNiconicoLive = "NL"
)

// Contents
//
// Contents in the Message API

// CtNagomeBroadOpen is a content of CommNagomeBroadOpen
type CtNagomeBroadOpen struct {
	Platform    string `json:"platform"`
	BroadID     string `json:"broad_id"`
	BroadURL    string `json:"broad_url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ChannelName string `json:"channel_name,omitempty"`
	ChannelID   string `json:"channel_id,omitempty"`
	ChannelURL  string `json:"channel_url,omitempty"`
	OwnerName   string `json:"owner_name,omitempty"`
	OwnerID     string `json:"owner_id,omitempty"`
	OwnerURL    string `json:"owner_url,omitempty"`
	OwnerBroad  bool   `json:"owner_broad,omitempty"`

	ScheduledStartTime time.Time `json:"scheduled_start_time,omitempty"`
	StartTime          time.Time `json:"start_time"`
	EndTime            time.Time `json:"end_time"`
}

// CtNagomeBroadInfo is a content of CommNagomeBroadInfo
type CtNagomeBroadInfo struct {
	Platform            string `json:"platform"`
	ViewCount           string `json:"view_count,omitempty"`
	ConcurrentViewCount string `json:"concurrent_view_count,omitempty"`
	CommentCount        string `json:"comment_count,omitempty"`
	LikeCount           string `json:"like_count,omitempty"`
	DislikeCount        string `json:"dislike_count,omitempty"`
}

type User struct {
	Platform     string    `json:"platform"`
	ID           string    `json:"id,omitempty"`
	Name         string    `json:"name"`
	CreateTime   time.Time `json:"create_time,omitempty"`
	Is184        bool      `json:"is184,omitempty"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
}

// CtNagomeUserUpdate is a content of CommNagomeUserUpdate
type CtNagomeUserUpdate User

// CtQueryBroadConnect is a content of CommQueryBroadConnect
type CtQueryBroadConnect struct {
	URL    string `json:"url"`
	RetryN int    `json:"retry_n,omitempty"` // Internal valuable that holds how many retries have done
}

// type of CtQueryBroadSendComment
const (
	CtQueryBroadSendCommentTypeGeneral string = "General"
	CtQueryBroadSendCommentTypeOwner          = "Owner" // ignored if the user is not the owner
)

// CtQueryBroadSendComment is a content of CommQueryBroadSendComment
type CtQueryBroadSendComment struct {
	Platform string `json:"platform"`
	Text     string `json:"text"`
	Iyayo    bool   `json:"iyayo,omitempty"`
	Type     string `json:"type,omitempty"` // if omitted, automatically selected depend on the settings
}

// CtQueryAccountSet is a content of CommQueryAccountSet
type CtQueryAccountSet struct {
	NicoLive *NicoLiveAccount `json:"nico_live,omitempty"`
}
type NicoLiveAccount struct {
	Mail        string `json:"mail"`
	Pass        string `json:"pass"`
	Usersession string `json:"usersession"`
}

// CtQueryLogPrint is a content of CommQueryLogPrint
type CtQueryLogPrint struct {
	Text string `json:"text"`
}

type SettingsSlot struct {
	Name            string          `yaml:"name"`
	AutoSaveTo0Slot bool            `yaml:"auto_save_to0_slot"`
	PluginDisable   map[string]bool `yaml:"plugin_disable"`

	Nicolive SettingsNicolive `yaml:"nicolive"`
}

type SettingsNicolive struct {
	UserNameGet  bool `yaml:"user_name_get"`
	OwnerComment bool `yaml:"owner_comment"`
}

// CtQuerySettingsSetCurrent is a content of CommQuerySettingsSetCurrent
type CtQuerySettingsSetCurrent SettingsSlot

type SettingsSlots struct {
	Config []*SettingsSlot `json:"config"`
}

// CtQuerySettingsSetAll is a content of CommQuerySettingsSetAll
type CtQuerySettingsSetAll SettingsSlots

// CtQueryPlugEnable is a content of CommQueryPlugEnable
type CtQueryPlugEnable struct {
	No     int  `json:"no"`
	Enable bool `json:"enable"`
}

// CtQueryUserSet is a content for CommQueryUserSet
type CtQueryUserSet User

// CtQueryUserSetName is a content for CommQueryUserSetName
type CtQueryUserSetName struct {
	Platform string `json:"platform"`
	ID       string `json:"id"`
	Name     string `json:"name"`
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
	Platform string    `json:"platform"`
	No       int       `json:"no"`
	Date     time.Time `json:"date"`
	Raw      string    `json:"raw"` // Untouched comment message.  Mainly for logging purpose.
	Comment  string    `json:"comment"`

	UserID           string `json:"user_id,omitempty"`
	UserName         string `json:"user_name"`
	UserThumbnailURL string `json:"user_thumbnail_url,omitempty"`
	UserUrl          string `json:"user_url,omitempty"`
	Score            int    `json:"score,omitempty"`
	IsPremium        bool   `json:"is_premium,omitempty"`
	IsBroadcaster    bool   `json:"is_broadcaster"`
	IsStaff          bool   `json:"is_staff,omitempty"`
	IsAnonymity      bool   `json:"is_anonymity"`
}

// CtUINotification is a content of CommUINotification
type CtUINotification struct {
	// Select type from below const string
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
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

// CtDirectUserGet is a content for CommDirectUserGet
type CtDirectUserGet struct {
	Platform string `json:"platform"`
	ID       string `json:"id"`
}

// CtDirectngmAppVersion is a content for CommDirectngmAppVersion
type CtDirectngmAppVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// CtDirectngmPlugList is a content for CommDirectngmPlugList
type CtDirectngmPlugList struct {
	Plugins []*Plugin `json:"plugins"`
}

// A Plugin is a Nagome plugin.
type Plugin struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Method      string   `json:"method"`
	Subscribe   []string `json:"subscribe"`
	No          int      `json:"no"`
	State       int      `json:"state"` // Don't change directly
}

// CtDirectngmSettingsCurrent is a content for CommDirectngmSettingsCurrent
type CtDirectngmSettingsCurrent SettingsSlot

// CtDirectngmSettingsAll is a content for CommDirectngmSettingsAll
type CtDirectngmSettingsAll SettingsSlots

// CtDirectngmUserGet is a content for CommDirectngmUserGet
type CtDirectngmUserGet User
