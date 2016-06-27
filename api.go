package main

import "encoding/json"

// Func names in NagomeMess
const (
	FuncnBroadQuery   string = "BroadQuery"
	FuncnAccountQuery        = "AccountQuery"
	FuncnComment             = "Comment"
)

// Index enum of BroadQuery
const (
	CommBroadQueryConnect     string = "Connect"
	CommBroadQueryDisconnect         = "Disconnect"
	CommBroadQuerySendComment        = "SendComment"
)

// Index enum of Account
const (
	CommAccountLogin string = "Login"
	CommAccountLoad         = "Load"
	CommAccountSave         = "Save"
)

// Index enum of Comment
const (
	CommCommentGot string = "Got"
)

// Message is base API struct for plugin
type Message struct {
	// Domain that includes following commands
	// ex. "Nagome"
	Domain string
	// Function
	//
	// Query
	Func string
	// Command
	Command string
	// Elements of Content is depend on Command
	Content json.RawMessage
}

// Contents
//
// Contents in the Message API

// BroadConnect requests to start receiving new broadcast
type BroadConnect struct {
	BroadID string
}
