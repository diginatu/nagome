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
	// Elements type of Content is depend on witch Command is used
	Content json.RawMessage
}

// NewMessage returns new Message with the given values.
func NewMessage(dom, fun, com string, con interface{}) (*Message, error) {
	conj, err := json.Marshal(con)
	if err != nil {
		return nil, err
	}

	m := &Message{
		Domain:  dom,
		Func:    fun,
		Command: com,
		Content: conj,
	}
	return m, nil
}

// Dimain names
const (
	DomainNagome string = "nagome"
)

// Func names
const (
	// Queries

	FuncQueryBroad   string = "QueryBroad"
	FuncQueryAccount        = "QueryAccount"

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
	CommQueryAccountSet   = "Set" // set the given content value as account values.
	CommQueryAccountLogin = "Login"
	CommQueryAccountLoad  = "Load"
	CommQueryAccountSave  = "Save"

	// Events

	// Comment
	CommCommentAdd = "Add"

	// Open
	CommUIDialog = "Dialog"
)

// Contents
//
// Contents in the Message API

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

// A CtCommentAdd is a content of got comment
type CtCommentAdd nicolive.Comment

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
