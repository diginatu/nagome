package nicolive

import (
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

	PostKey           string
	OwnerBroad        bool
	OwnerCommentToken string
}

// IsUserOwner returns whether the user own this broad
func (l *LiveWaku) IsUserOwner() bool {
	return l.Stream.OwnerID == l.User.UserID
}

// FetchInformation gets information using getplayerstatus API
func (l *LiveWaku) FetchInformation() NicoError {
	if l.Account == nil {
		return NicoErr(NicoErrOther, "no account",
			"LiveWaku does not have an account")
	}
	if l.BroadID == "" {
		return NicoErr(NicoErrOther, "no BroadID",
			"BroadID is not set")
	}

	c, nicoerr := NewNicoClient(l.Account)
	if nicoerr != nil {
		return nicoerr
	}

	url := fmt.Sprintf(
		"http://watch.live.nicovideo.jp/api/getplayerstatus/%s",
		l.BroadID)
	res, err := c.Get(url)
	if err != nil {
		return NicoErrFromStdErr(err)
	}
	defer res.Body.Close()

	root, err := xmlpath.Parse(res.Body)
	if err != nil {
		return NicoErrFromStdErr(err)
	}

	if v, ok := statusXMLPath.String(root); ok {
		if v != "ok" {
			if v, ok := errorCodeXMLPath.String(root); ok {
				errorNum := NicoErrNicoLiveOther
				switch v {
				case "closed":
					errorNum = NicoErrClosed
				case "notlogin":
					errorNum = NicoErrNotLogin
				}
				return NicoErr(errorNum, v, "")
			}
			return NicoErr(NicoErrOther, "FetchInformation unknown err",
				"request failed with unknown error")
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
func (l *LiveWaku) FetchHeartBeat() (HeartbeatValue, NicoError) {
	var hb HeartbeatValue

	if l.Account == nil {
		return hb, NicoErr(NicoErrOther, "no account", "LiveWaku does not have an account")
	}
	if l.BroadID == "" {
		return hb, NicoErr(NicoErrOther, "no BroadID", "BroadID is not set")
	}

	c, nicoerr := NewNicoClient(l.Account)
	if nicoerr != nil {
		return hb, nicoerr
	}

	url := fmt.Sprintf(
		"http://live.nicovideo.jp/api/heartbeat?v=%s",
		l.BroadID)
	res, err := c.Get(url)
	if err != nil {
		return hb, NicoErrFromStdErr(err)
	}
	defer res.Body.Close()

	root, err := xmlpath.Parse(res.Body)
	if err != nil {
		return hb, NicoErrFromStdErr(err)
	}

	if v, ok := statusXMLPath.String(root); ok {
		if v != "ok" {
			if v, ok := errorCodeXMLPath.String(root); ok {
				errorNum := NicoErrNicoLiveOther
				if v == "NOTLOGIN" {
					errorNum = NicoErrNotLogin
				}

				var desc string
				if v, ok := errorDescXMLPath.String(root); ok {
					desc = v
				}
				return hb, NicoErr(errorNum, v, desc)
			}
			return hb, NicoErr(NicoErrOther, "unknown err", "request failed with unknown error")
		}
	}

	// stream
	if v, ok := xmlpath.MustCompile("/heartbeat/watchCount").String(root); ok {
		hb.WatchCount = v
	}
	if v, ok := xmlpath.MustCompile("/heartbeat/commentCount").String(root); ok {
		hb.CommentCount = v
	}

	if hb.CommentCount == "" || hb.WatchCount == "" {
		return hb, NicoErr(NicoErrOther, "unknown err", "could not get")
	}

	return hb, nil
}
