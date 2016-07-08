package main

import "encoding/json"

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

// Func names in NagomeMess
// Query suffix means the Func is Query, which ask Nagome to do something.
// Other Funcs are all Event.
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

// Contents
//
// Contents in the Message API

// CtBroadQueryConnect requests to start receiving new broadcast
type CtBroadQueryConnect struct {
	BroadID string
}

// CtBroadQuerySendComment requests to send comment
type CtBroadQuerySendComment struct {
	Text  string
	Iyayo bool
}
