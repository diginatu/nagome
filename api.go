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

// Dimain names
const (
	DomainNagome string = "nagome"
)

// Func names
const (
	// Queries
	FuncQueryBroad   string = "BroadQuery"
	FuncQueryAccount        = "AccountQuery"

	// Events
	FuncComment = "Comment"
)

// Command names
const (
	// QueryBroad
	CommQueryBroadConnect     string = "Connect"
	CommQueryBroadDisconnect         = "Disconnect"
	CommQueryBroadSendComment        = "SendComment"

	// QueryAccount
	CommQueryAccountLogin string = "Login"
	CommQueryAccountLoad         = "Load"
	CommQueryAccountSave         = "Save"

	// Comment
	CommCommentGot string = "Got"
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
