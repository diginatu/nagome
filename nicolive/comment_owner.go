package nicolive

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
)

// CommentOwner sends a comment as the owner.
func CommentOwner(lw *LiveWaku, text, name string, ac *Account) error {
	if lw.OwnerCommentToken == "" {
		return MakeError(ErrOther, "empty token")
	}

	v := url.Values{}
	v.Add("body", text)
	v.Add("token", lw.OwnerCommentToken)
	v.Add("name", name)

	urls := fmt.Sprintf("http://watch.live.nicovideo.jp/api/broadcast/%s?%s",
		lw.BroadID, v.Encode())

	return commentOwnerImpl(urls, ac)
}

func commentOwnerImpl(urls string, ac *Account) (err error) {
	res, err := ac.client.Get(urls)
	if err != nil {
		return MakeError(ErrSendComment, err.Error())
	}
	defer func() {
		if lerr := res.Body.Close(); lerr != nil {
			if err == nil {
				err = lerr
			}
		}
	}()

	brs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return MakeError(ErrSendComment, err.Error())
	}

	v, err := url.ParseQuery(string(brs))
	if err != nil {
		return MakeError(ErrSendComment, err.Error())
	}

	if len(v["status"]) > 0 {
		if v["status"][0] == "error" {
			if len(v["error"]) > 0 {
				return MakeError(ErrSendComment, "error num : "+v["error"][0])
			}
			return MakeError(ErrSendComment, "unknown error")
		}
		return nil
	}

	fmt.Fprintln(os.Stderr, v)
	return MakeError(ErrSendComment, "unknown error")
}
