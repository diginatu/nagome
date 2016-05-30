package main

import "encoding/json"

// Index enum of NagomeMess
const (
	FuncnBroadQuery int = iota
	FuncnAccountQuery
	FuncnComment
)

// Index enum of BroadQuery
const (
	CommBroadQueryConnect int = iota
	CommBroadQueryDisconnect
	CommBroadQuerySendComment
)

// Index enum of Account
const (
	CommAccountLogin int = iota
	CommAccountLoad
	CommAccountSave
)

// Index enum of Comment
const (
	CommCommentGot int = iota
)

var (
	// NagomeMess holds possible Funcs and commands in Message of Nagome Domain.
	NagomeMess = [3]struct {
		Funcn    string
		Commands []string
	}{
		// Query (Plugin to Nagome)
		{
			"BroadQuery",
			[]string{"Connect", "Disconnect", "SendComment"},
		},
		{
			"AccountQuery",
			[]string{"Login", "Load", "Save"},
		},
		// Event
		{
			"Comment",
			[]string{"Got"},
		},
	}
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
	// Elements of Content is depend on Domain, Func and Command
	Content json.RawMessage
}

// Contents
//
// Contents in the Message API

// BroadConnect requests to start receiving new broadcast
type BroadConnect struct {
	BroadID string
}
