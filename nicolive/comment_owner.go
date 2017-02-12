package nicolive

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// CommentOwner sends a comment as the owner.
func CommentOwner(lw *LiveWaku, text, name string) error {
	if lw.OwnerCommentToken == "" {
		return MakeError(ErrOther, "empty token")
	}

	v := url.Values{}
	v.Add("body", text)
	v.Add("token", lw.OwnerCommentToken)
	v.Add("name", name)

	urls := fmt.Sprintf("http://watch.live.nicovideo.jp/api/broadcast/%s?%s",
		lw.BroadID, v.Encode())

	return commentOwnerImpl(urls)
}

func commentOwnerImpl(urls string) (err error) {
	c := http.Client{}
	res, err := c.Get(urls)
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

	fmt.Println()
	brs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return MakeError(ErrSendComment, err.Error())
	}

	v, err := url.ParseQuery(string(brs))
	if err != nil {
		return MakeError(ErrSendComment, err.Error())
	}

	fmt.Println(v)

	return nil
}
