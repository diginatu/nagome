package main

import (
	"encoding/json"

	"github.com/diginatu/nagome/nicolive"
)

// Message is base API struct for plugin
type Message struct {
	// Domain that includes following parameters
	Domain string
	// Function
	Func string
	// Command
	Command string
	// Elements of Content is depend on Command
	Content json.RawMessage
}

// NewMessage returns new Message with the given values.
func NewMessage(dom, fun, com string, con interface{}) (m *Message, err error) {
	conj, err := json.Marshal(con)

	m = &Message{
		Domain:  dom,
		Func:    fun,
		Command: com,
		Content: conj,
	}
	return
}

// Dimain names
const (
	DomainNagome string = "Nagome"
)

// Func names
const (
	// Queries

	FuncQueryBroad   string = "BroadQuery"
	FuncQueryAccount        = "AccountQuery"

	// Events

	// FuncComment is about a comment connection for an account and a broadcast
	FuncComment = "Comment"

	// FuncOpen is an event (request) for UI
	FuncUI = "UI"
)

// Command names
const (
	// Queries

	// QueryBroad
	CommQueryBroadConnect     string = "Connect"
	CommQueryBroadDisconnect         = "Disconnect"
	CommQueryBroadSendComment        = "SendComment"

	// QueryAccount
	CommQueryAccountLogin = "Login"
	CommQueryAccountLoad  = "Load"
	CommQueryAccountSave  = "Save"

	// Events

	// Comment
	CommAddComment = "Got"

	// Open
	CommUIDialog = "Dialog"
)

// Contents
//
// Contents in the Message API

// CtQueryBroadConnect requests to start receiving new broadcast
type CtQueryBroadConnect struct {
	BroadID string
}

// CtQueryBroadSendComment requests to send comment
type CtQueryBroadSendComment struct {
	Text  string
	Iyayo bool
}

// A CtCommentGot is a content of got comment
type CtCommentGot nicolive.Comment

// CtUIDialog is content of dialog that nagome ask to open
type CtUIDialog struct {
	// Type value is "info" or "warn"
	Type        string
	Title       string
	Description string
}
