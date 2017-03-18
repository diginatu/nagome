package nicolive

import (
	"encoding/xml"
	"fmt"
)

// PublishStatusItem is the response of PublishStatus API
type PublishStatusItem struct {
	Token   string `json:"token"`
	URL     string `json:"url"`
	Stream  string `json:"stream"`
	Ticket  string `json:"ticket"`
	Bitrate string `json:"bitrate"`
}

// PublishStatus gets a token to comment as owner.
// This function is safe for concurrent use.
func PublishStatus(broadID string, a *Account) (*PublishStatusItem, error) {
	return publishStatusImpl(
		fmt.Sprintf("http://live.nicovideo.jp/api/getpublishstatus?v=%s", broadID), a)
}

func publishStatusImpl(url string, a *Account) (ps *PublishStatusItem, err error) {
	type pbsxml struct {
		Status  string `xml:"status,attr"`
		Code    string `xml:"error>code"`
		Token   string `xml:"stream>token"`
		URL     string `xml:"rtmp>url"`
		Stream  string `xml:"rtmp>stream"`
		Ticket  string `xml:"rtmp>ticket"`
		Bitrate string `xml:"rtmp>bitrate"`
	}

	cl := a.client
	if cl == nil {
		return nil, MakeError(ErrOther, "nil account http client")
	}

	res, err := cl.Get(url)
	if err != nil {
		return nil, ErrFromStdErr(err)
	}
	defer func() {
		if lerr := res.Body.Close(); lerr != nil && err == nil {
			err = lerr
		}
	}()

	r := new(pbsxml)
	dec := xml.NewDecoder(res.Body)
	err = dec.Decode(r)
	if err != nil {
		return nil, ErrFromStdErr(err)
	}

	if r.Status == "fail" {
		return nil, MakeError(ErrOther, "PublishStatus failed : "+r.Code)
	}

	return &PublishStatusItem{
		Bitrate: r.Bitrate,
		Stream:  r.Stream,
		Ticket:  r.Ticket,
		Token:   r.Token,
		URL:     r.URL,
	}, nil
}
