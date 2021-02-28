package api

import (
	"time"
)

// Dimain names
const (
	DomainNicolive = "nagome_nicolive"
)

type user struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	CreateTime   time.Time `json:"create_time"`
	Is184        bool      `json:"is184"`
	ThumbnailURL string    `json:"thumbnail_url"`
}

// CtNagomeUserUpdate is a content of CommNagomeUserUpdate
type CtNagomeUserUpdate user

// CtQueryAccountSet is a content of CommQueryAccountSet
type CtQueryAccountSet struct {
	Mail        string `json:"mail"`
	Pass        string `json:"pass"`
	Usersession string `json:"usersession"`
}

// CtQueryUserSet is a content for CommQueryUserSet
type CtQueryUserSet user

// CtDirectngmUserGet is a content for CommDirectngmUserGet
type CtDirectngmUserGet user
