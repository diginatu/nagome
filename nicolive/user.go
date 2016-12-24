package nicolive

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"gopkg.in/xmlpath.v2"
)

const (
	userDBDirName = "userdb"
)

// User is a niconico user.
type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	GotTime      time.Time `json:"got_time"`
	Is184        bool      `json:"is184"`
	ThumbnailURL string    `json:"thumbnail_url"`
}

// FetchUserInfo fetches user name and Thumbnail URL from niconico.
func FetchUserInfo(id string, a *Account) (*User, error) {
	url := fmt.Sprintf("http://api.ce.nicovideo.jp/api/v1/user.info?user_id=%s", id)
	return fetchUserInfoImpl(url, a)
}

func fetchUserInfoImpl(url string, a *Account) (user *User, err error) {
	u := new(User)

	c, nerr := NewNicoClient(a)
	if nerr != nil {
		return nil, nerr
	}

	res, err := c.Get(url)
	if err != nil {
		return nil, ErrFromStdErr(err)
	}
	defer func() {
		if lerr := res.Body.Close(); lerr != nil {
			if err == nil {
				err = lerr
			}
		}
	}()

	root, err := xmlpath.Parse(res.Body)
	if err != nil {
		return nil, ErrFromStdErr(err)
	}

	if v, ok := xmlPathStatus.String(root); ok {
		if v != "ok" {
			if v, ok := xmlPathErrorCode.String(root); ok {
				desc, _ := xmlPathErrorDesc.String(root)
				return nil, MakeError(ErrOther, v+desc)
			}
			return nil, MakeError(ErrOther, "request failed with unknown error")
		}
	}

	// stream
	if v, ok := xmlpath.MustCompile("/nicovideo_user_response/user/id").String(root); ok {
		u.ID = v
	}
	if v, ok := xmlpath.MustCompile("/nicovideo_user_response/user/nickname").String(root); ok {
		u.Name = v
	}
	if v, ok := xmlpath.MustCompile("/nicovideo_user_response/user/thumbnail_url").String(root); ok {
		u.ThumbnailURL = v
	}

	u.GotTime = time.Now()

	return u, nil
}

// Equal reports whether t and x represent the same User instant.
func (u *User) Equal(x *User) bool {
	return u.ID == x.ID &&
		u.Name == x.Name &&
		u.GotTime.Unix() == x.GotTime.Unix() &&
		u.Is184 == x.Is184 &&
		u.ThumbnailURL == x.ThumbnailURL
}

// UserDB is database of Users.
type UserDB struct {
	db *leveldb.DB
}

// NewUserDB creates new UserDB.
func NewUserDB(dirname string) (*UserDB, error) {
	db, err := leveldb.OpenFile(dirname, nil)
	if err != nil {
		return nil, ErrFromStdErr(err)
	}

	return &UserDB{db}, nil
}

// Store stores a user into the DB.
func (d *UserDB) Store(u *User) error {
	b, err := json.Marshal(u)
	if err != nil {
		return ErrFromStdErr(err)
	}
	err = d.db.Put([]byte(u.ID), b, nil)
	if err != nil {
		return ErrFromStdErr(err)
	}
	return nil
}

// Fetch fetches a user of given ID from the DB.
// If no user is found, return (nil, nil).
// So you should check if user is nil.
func (d *UserDB) Fetch(id string) (*User, error) {
	var u = new(User)
	b, err := d.db.Get([]byte(id), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, ErrFromStdErr(err)
	}
	err = json.Unmarshal(b, u)
	if err != nil {
		return nil, ErrFromStdErr(err)
	}
	return u, nil
}

// Remove removes a user of given ID from the DB.
func (d *UserDB) Remove(id string) error {
	err := d.db.Delete([]byte(id), nil)
	if err != nil {
		return ErrFromStdErr(err)
	}
	return nil
}

// Close closes the DB.
func (d *UserDB) Close() error {
	err := d.db.Close()
	if err != nil {
		return err
	}
	return nil
}
