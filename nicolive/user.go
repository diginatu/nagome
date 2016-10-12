package nicolive

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // sqlite3 for database/sql
	"gopkg.in/xmlpath.v2"
)

// User is a niconico user.
type User struct {
	ID           string
	Name         string
	GotTime      time.Time // The information shorter than 1 second will be lost after storing to UserDB.
	Is184        bool
	ThumbnailURL string
	Misc         string
}

// FetchInfo fetches user name and Thumbnail URL from niconico.
func (u *User) FetchInfo(id string, a *Account) Error {
	url := fmt.Sprintf("http://api.ce.nicovideo.jp/api/v1/user.info?user_id=%s", id)
	return u.fetchInfoImpl(url, a)
}

func (u *User) fetchInfoImpl(url string, a *Account) Error {
	c, nerr := NewNicoClient(a)
	if nerr != nil {
		return nerr
	}

	res, err := c.Get(url)
	if err != nil {
		return ErrFromStdErr(err)
	}
	defer res.Body.Close()

	root, err := xmlpath.Parse(res.Body)
	if err != nil {
		return ErrFromStdErr(err)
	}

	if v, ok := statusXMLPath.String(root); ok {
		if v != "ok" {
			if v, ok := errorCodeXMLPath.String(root); ok {
				desc, _ := errorDescXMLPath.String(root)
				return MakeError(ErrOther, v+desc)
			}
			return MakeError(ErrOther, "request failed with unknown error")
		}
	}

	// stream
	if v, ok := xmlpath.MustCompile("/nicovideo_user_response/user/nickname").String(root); ok {
		u.Name = v
	}
	if v, ok := xmlpath.MustCompile("/nicovideo_user_response/user/thumbnail_url").String(root); ok {
		u.ThumbnailURL = v
	}

	return nil
}

// Equal reports whether t and x represent the same User instant.
func (u *User) Equal(x *User) bool {
	return u.ID == x.ID &&
		u.Name == x.Name &&
		u.GotTime.Unix() == x.GotTime.Unix() &&
		u.Is184 == x.Is184 &&
		u.ThumbnailURL == x.ThumbnailURL &&
		u.Misc == x.Misc
}

const (
	userDBCreateUserTable = `
create table if not exists user (
	id text primary key,
	name text,
	got_time integer,
	is_184 integer,
	thumbnail_url text,
	misc text
)`
	userDBInsertUser = `
insert or replace into
user (id, name, got_time, is_184, thumbnail_url, misc)
values (?, ?, ?, ?, ?, ?)`
	userDBFetchUser  = `select * from user where id=?`
	userDBRemoveUser = `delete from user where id=?`
)

// UserDB is database of Users.
type UserDB struct {
	File      string
	db        *sql.DB
	mu        sync.RWMutex
	fetchStmt *sql.Stmt
}

// NewUserDB creates new UserDB.
func NewUserDB(file string) (*UserDB, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(userDBCreateUserTable)
	if err != nil {
		return nil, err
	}

	fstmt, err := db.Prepare(userDBFetchUser)
	if err != nil {
		return nil, err
	}

	return &UserDB{
		File:      file,
		db:        db,
		fetchStmt: fstmt,
	}, nil
}

// Store stores a user into the DB.
func (d *UserDB) Store(u *User) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(userDBInsertUser, u.ID, u.Name, u.GotTime.Unix(), u.Is184, u.ThumbnailURL, u.Misc)
	if err != nil {
		return err
	}

	return nil
}

// Fetch fetches a user of given ID from the DB.
// If no user is found, return (nil, nil)
func (d *UserDB) Fetch(id string) (*User, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var u User
	var t int64
	err := d.fetchStmt.QueryRow(id).Scan(&u.ID, &u.Name, &t, &u.Is184, &u.ThumbnailURL, &u.Misc)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	u.GotTime = time.Unix(t, 0)

	return &u, nil
}

// Remove removes a user of given ID from the DB.
func (d *UserDB) Remove(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(userDBRemoveUser, id)
	if err != nil {
		return err
	}

	return nil
}

// Close closes the DB.
func (d *UserDB) Close() {
	d.fetchStmt.Close()
	d.db.Close()
}
