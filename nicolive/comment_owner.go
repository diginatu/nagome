package nicolive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type CommentOwnerRequest struct {
	Text        string `json:"text"`
	IsPermanent bool   `json:"isPermanent"`
	Color       string `json:"color"`
	UserName    string `json:"userName,omitempty"`
}

// CommentOwner sends a comment as the owner.
func CommentOwner(broadID string, method string, commreq *CommentOwnerRequest, ac *Account) error {
	url := fmt.Sprintf("http://live2.nicovideo.jp/watch/%s/operator_comment", broadID)

	return commentOwnerImpl(method, commreq, url, ac)
}

func commentOwnerImpl(method string, commreq *CommentOwnerRequest, url string, ac *Account) (err error) {
	type CommentOwnerResponse struct {
		Meta struct {
			ErrorCode    string `json:"errorCode"`
			Status       int    `json:"status"`
			ErrorMessage string `json:"errorMessage"`
		} `json:"meta"`
	}

	commreqBytes, err := json.Marshal(commreq)
	if err != nil {
		return MakeError(ErrSendComment, err.Error())
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(commreqBytes))
	if err != nil {
		return MakeError(ErrSendComment, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := ac.client.Do(req)
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

	commResp := new(CommentOwnerResponse)
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(commResp)
	if err != nil {
		return MakeError(ErrSendComment, err.Error())
	}

	if commResp.Meta.Status != 200 {
		return MakeError(ErrSendComment, fmt.Sprintf("failed to send owner comment with an error Status: %d Code: %s Message: %s", commResp.Meta.Status, commResp.Meta.ErrorCode, commResp.Meta.ErrorMessage))
	}

	return nil
}
