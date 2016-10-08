package nicolive

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	userInfoResponseOkName = "テスト名前"
	userInfoResponseOkThum = "testurl"
	userInfoResponseOk     = `<?xml version="1.0" encoding="UTF-8"?>
<nicovideo_user_response status="ok">
  <user>
    <id>1</id>
    <nickname>` + userInfoResponseOkName + `</nickname>
    <thumbnail_url>` + userInfoResponseOkThum + `</thumbnail_url>
  </user>
  <vita_option>
    <user_secret>0</user_secret>
  </vita_option>
  <additionals/>
</nicovideo_user_response>`
	userInfoResponseNotFound = `<?xml version="1.0" encoding="UTF-8"?>
<nicovideo_user_response status="fail"><error><code>NOT_FOUND</code><description>ユーザーが見つかりません</description></error></nicovideo_user_response>`
)

func TestUserFetchInfo(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, userInfoResponseOk)
		},
	)
	mux.HandleFunc("/notfound",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, userInfoResponseNotFound)
		},
	)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	var u User
	a := &Account{Usersession: "usersession_example"}

	if nerr := u.fetchInfoImpl(ts.URL+"/ok", a); nerr != nil {
		log.Fatal(nerr)
	}
	if u.Name != userInfoResponseOkName {
		t.Fatalf("Should be %v but %v", userInfoResponseOkName, u.Name)
	}
	if u.ThumbnailURL != userInfoResponseOkThum {
		t.Fatalf("Should be %v but %v", userInfoResponseOkThum, u.ThumbnailURL)
	}

	if nerr := u.fetchInfoImpl(ts.URL+"/notfound", a); nerr == nil {
		log.Fatal(nerr)
	}
}

func TestUserDB(t *testing.T) {
	f, err := ioutil.TempFile("", "nagome")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	db, err := NewUserDB(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
}
