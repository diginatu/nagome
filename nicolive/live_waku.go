package nicolive

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/xmlpath.v2"
)

// HeartbeatValue is struct to hold result of heartbeat API
type HeartbeatValue struct {
	WatchCount   string
	CommentCount string
}

// LiveWaku is a live broadcast(Waku) of Niconama
type LiveWaku struct {
	Account *Account
	BroadID string

	Stream struct {
		Title       string
		Description string
		CommunityID string
		OwnerID     string
		OwnerName   string

		OpenTime  time.Time
		StartTime time.Time
		EndTime   time.Time

		BroadcastToken string
	}

	User struct {
		UserID    string
		Name      string
		IsPremium bool
	}

	CommentServer struct {
		Addr   string
		Port   string
		Thread string
	}

	OwnerBroad        bool
	OwnerCommentToken string
}

// IsUserOwner returns whether the user own this broad
func (l *LiveWaku) IsUserOwner() bool {
	return l.Stream.OwnerID == l.User.UserID
}

// FetchInformation gets information using getplayerstatus API
func (l *LiveWaku) FetchInformation() (err error) {
	if l.Account == nil {
		return MakeError(ErrIncorrectAccount, "nil Account in LiveWaku")
	}
	if l.BroadID == "" {
		return MakeError(ErrOther, "BroadID is not set")
	}

	c := l.Account.client
	if c == nil {
		return MakeError(ErrOther, "nil Account http client in LiveWaku")
	}

	url := fmt.Sprintf("http://watch.live.nicovideo.jp/api/getplayerstatus/%s", l.BroadID)
	res, err := c.Get(url)
	if err != nil {
		return MakeError(ErrNetwork, "client.Get : "+err.Error())
	}
	defer func() {
		lerr := res.Body.Close()
		if lerr != nil && err == nil {
			err = lerr
		}
	}()

	root, err := xmlpath.Parse(res.Body)
	if err != nil {
		return ErrFromStdErr(err)
	}

	if v, ok := xmlPathStatus.String(root); ok {
		if v != "ok" {
			if v, ok := xmlPathErrorCode.String(root); ok {
				errorNum := ErrNicoLiveOther
				switch v {
				case "closed":
					errorNum = ErrClosed
				case "notlogin":
					errorNum = ErrNotLogin
				}
				return MakeError(errorNum, v)
			}
			return MakeError(ErrOther, "FetchInformation failed with unknown err")
		}
	}

	// stream
	if v, ok := xmlpath.MustCompile("//stream/title").String(root); ok {
		l.Stream.Title = v
	}
	if v, ok := xmlpath.MustCompile("//stream/description").String(root); ok {
		l.Stream.Description = v
	}
	if v, ok := xmlpath.MustCompile("//stream/default_community").String(root); ok {
		l.Stream.CommunityID = v
	}
	if v, ok := xmlpath.MustCompile("//stream/owner_id").String(root); ok {
		l.Stream.OwnerID = v
	}
	if v, ok := xmlpath.MustCompile("//stream/owner_name").String(root); ok {
		l.Stream.OwnerName = v
	}
	if v, ok := xmlpath.MustCompile("//stream/open_time").String(root); ok {
		i, _ := strconv.Atoi(v)
		l.Stream.OpenTime = time.Unix(int64(i), 0)
	}
	if v, ok := xmlpath.MustCompile("//stream/start_time").String(root); ok {
		i, _ := strconv.Atoi(v)
		l.Stream.StartTime = time.Unix(int64(i), 0)
	}
	if v, ok := xmlpath.MustCompile("//stream/end_time").String(root); ok {
		i, _ := strconv.Atoi(v)
		l.Stream.EndTime = time.Unix(int64(i), 0)
	}
	if v, ok := xmlpath.MustCompile("//stream/broadcast_token").String(root); ok {
		l.Stream.BroadcastToken = v
	}

	// user
	if v, ok := xmlpath.MustCompile("//user/user_id").String(root); ok {
		l.User.UserID = v
	}
	if v, ok := xmlpath.MustCompile("//user/nickname").String(root); ok {
		l.User.Name = v
	}
	if v, ok := xmlpath.MustCompile("//user/is_premium").String(root); ok {
		l.User.IsPremium, _ = strconv.ParseBool(v)
	}

	// comment server
	if v, ok := xmlpath.MustCompile("//ms/addr").String(root); ok {
		l.CommentServer.Addr = v
	}
	if v, ok := xmlpath.MustCompile("//ms/port").String(root); ok {
		l.CommentServer.Port = v
	}
	if v, ok := xmlpath.MustCompile("//ms/thread").String(root); ok {
		l.CommentServer.Thread = v
	}

	return nil
}

// FetchHeartBeat gets watcher and comment count using heartbeat API
// This function is safe for concurrent use.
func (l *LiveWaku) FetchHeartBeat() (heartBeatValue *HeartbeatValue, waitTime int, err error) {
	if l.Account == nil {
		return nil, 0, MakeError(ErrOther, "nil account in LiveWaku")
	}
	if l.BroadID == "" {
		return nil, 0, MakeError(ErrOther, "BroadID is not set")
	}

	type heartbeatXML struct {
		Status       string `xml:"status,attr"`
		Code         string `xml:"error>code"`
		Description  string `xml:"error>description"`
		WatchCount   string `xml:"watchCount"`
		CommentCount string `xml:"commentCount"`
		WaitTime     int    `xml:"waitTime"`
	}

	c := l.Account.client
	if c == nil {
		return nil, 0, MakeError(ErrOther, "nil account client")
	}

	url := fmt.Sprintf("http://live.nicovideo.jp/api/heartbeat?v=%s", l.BroadID)
	res, err := c.Get(url)
	if err != nil {
		return nil, 0, ErrFromStdErr(err)
	}
	defer func() {
		lerr := res.Body.Close()
		if lerr != nil && err == nil {
			err = lerr
		}
	}()

	r := new(heartbeatXML)
	dec := xml.NewDecoder(res.Body)
	err = dec.Decode(r)
	if err != nil {
		return nil, 0, ErrFromStdErr(err)
	}

	if r.Status != "ok" {
		errorNum := ErrNicoLiveOther
		if r.Code == "NOTLOGIN" {
			errorNum = ErrNotLogin
		}
		return nil, 0, MakeError(errorNum, "HeartbeatStatus failed : "+r.Code+" : "+r.Description)
	}

	if r.CommentCount == "" || r.WatchCount == "" {
		return nil, 0, MakeError(ErrOther, "heartbeat : unknown err")
	}

	return &HeartbeatValue{
		WatchCount:   r.WatchCount,
		CommentCount: r.CommentCount,
	}, r.WaitTime, nil
}
