package nicolive

import (
	"encoding/json"
	"fmt"
	"time"
	"unicode"

	"github.com/diginatu/nagome/api"
	"github.com/syndtr/goleveldb/leveldb"
	"gopkg.in/xmlpath.v2"
)

const (
	userDBDirName = "userdb"
)

// Is184UserID returns whether the ID is 184
func Is184UserID(id string) bool {
	for _, c := range id {
		if !unicode.IsDigit(c) {
			return true
		}
	}
	return false
}

// User is a niconico user.
type User struct {
	ID           string
	Name         string
	CreateTime   time.Time
	Is184        bool
	ThumbnailURL string
}

// CreateUser gather the user infomation of the given user id and returns pointer to new User struct.
func CreateUser(id string, a *Account) (*User, error) {
	if Is184UserID(id) {
		return &User{
			ID:    id,
			Is184: true,
		}, nil
	}

	u, err := FetchUserInfo(id, a)
	if err != nil {
		return nil, err
	}
	u.CreateTime = time.Now()
	return u, nil
}

// FetchUserInfo fetches user name and Thumbnail URL from niconico.
// This function is safe for concurrent use.
func FetchUserInfo(id string, a *Account) (*User, error) {
	url := fmt.Sprintf("http://api.ce.nicovideo.jp/api/v1/user.info?user_id=%s", id)
	return fetchUserInfoImpl(url, a)
}

func fetchUserInfoImpl(url string, a *Account) (user *User, err error) {
	u := new(User)

	c := a.client
	if c == nil {
		return nil, MakeError(ErrOther, "nil account http client")
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

	return u, nil
}

// Equal reports whether t and x represent the same User instant.
func (u *User) Equal(x *User) bool {
	return u.ID == x.ID &&
		u.Name == x.Name &&
		u.CreateTime.Unix() == x.CreateTime.Unix() &&
		u.Is184 == x.Is184 &&
		u.ThumbnailURL == x.ThumbnailURL
}

// API creates new API representation of the User
func (u *User) API() (x *api.User) {
	return &api.User{
		Platform:     api.PlatformNiconicoLive,
		ID:           u.ID,
		Name:         u.Name,
		CreateTime:   u.CreateTime,
		Is184:        u.Is184,
		ThumbnailURL: u.ThumbnailURL,
	}
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
func (d *UserDB) Fetch(id string) (*User, error) {
	var u = new(User)
	b, err := d.db.Get([]byte(id), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, MakeError(ErrDBUserNotFound, err.Error())
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
	return d.db.Close()
}
