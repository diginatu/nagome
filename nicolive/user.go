package nicolive

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3" // sqlite3 for database/sql
	"gopkg.in/xmlpath.v2"
)

// User is a niconico user
type User struct {
	ID           string
	Name         string
	GotTime      time.Time
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

// UserDB is database of Users.
type UserDB struct {
	File string
	DB   *sql.DB
}

// NewUserDB creates new UserDB.
func NewUserDB(file string) (*UserDB, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	stmt, err := db.Prepare("create table if not exists user " +
		"(id integer unique primary key, name varchar(60))")
	if err != nil {
		return nil, err
	}

	res, err := stmt.Exec()
	if err != nil {
		return nil, err
	}

	fmt.Println(res)

	return &UserDB{
		File: file,
		DB:   db,
	}, nil
}

// Store stores a user into the DB.
func (d *UserDB) Store(u User) error {
	return nil
}

// Fetch fetches a user of given ID from the DB.
func (d *UserDB) Fetch(id string) (*User, error) {
	return nil, nil
}

// Close closes the DB.
func (d *UserDB) Close() {
	d.DB.Close()
}
