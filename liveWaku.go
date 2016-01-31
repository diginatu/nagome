package nicolive

import (
	"fmt"
	"strconv"
	"time"

	"gopkg.in/xmlpath.v1"
)

var (
	statusPath    = xmlpath.MustCompile("//@status")
	errorCodePath = xmlpath.MustCompile("//error/code")
)

// LiveWaku is a live broadcast(Waku) of Niconama
type LiveWaku struct {
	Account *Account

	Stream struct {
		BroadID     string
		Title       string
		Description string
		CommunityID string
		OwnerID     string
		OwnerName   string

		StartTime time.Time
		EndTime   time.Time

		BroadcastToken string
	}

	User struct {
		UserID    string
		Name      string
		IsPremium int
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

// FetchInformation gets information using PlayerStatus API
func (l *LiveWaku) FetchInformation() error {
	if l.Account == nil {
		return fmt.Errorf("account is a nil value")
	}

	c, err := NewNicoClient(l.Account)
	if err != nil {
		return err
	}

	if l.Stream.BroadID == "" {
		return fmt.Errorf("Stream.BroadID is not set")
	}
	u := fmt.Sprintf(
		"http://watch.live.nicovideo.jp/api/getplayerstatus?v=%s",
		l.Stream.BroadID)

	res, err := c.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	root, err := xmlpath.Parse(res.Body)
	if err != nil {
		return err
	}

	if v, ok := statusPath.String(root); ok {
		if v != "ok" {
			if v, ok := errorCodePath.String(root); ok {
				return fmt.Errorf("returned error with code : %s", v)
			}
			return fmt.Errorf("returned error")
		}
	}

	// stream
	if v, ok := xmlpath.MustCompile("//stream/id").String(root); ok {
		l.Stream.BroadID = v
	}
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
		i, _ := strconv.Atoi(v)
		l.User.IsPremium = i
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

	fmt.Println(l)

	return nil
}
