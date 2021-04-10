package nicolive

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestPublishStatus(t *testing.T) {
	var (
		publishStatusv1 = &PublishStatusItem{
			Token:   "1c3e0fd476bde20a69999f8gsd7h98s7fgh97",
			URL:     "rtmp://nlpoca129.live.nicovideo.jp:1935/publicorigin/170202_00_0",
			Stream:  "lv289355780",
			Ticket:  "11246304:lv444444440:4:1484444435:0:1485964489:0aa54aea811d2cc9",
			Bitrate: "384",
		}

		publishStatus1 = fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<getpublishstatus status="ok" time="1485966235"><stream><id>lv289355780</id><token>%s</token><exclude>0</exclude><provider_type>community</provider_type><base_time>1485964486</base_time><open_time>1485964486</open_time><start_time>1485964489</start_time><end_time>1485966289</end_time><allow_vote>0</allow_vote><disable_adaptive_bitrate>1</disable_adaptive_bitrate><is_reserved>0</is_reserved><is_chtest>0</is_chtest><for_mobile>1</for_mobile><editstream_language>1</editstream_language><test_extend_enabled>1</test_extend_enabled><category>一般(その他)</category></stream><user><nickname>デジネイ</nickname><is_premium>1</is_premium><user_id>11246304</user_id><NLE>1</NLE></user><rtmp is_fms="1"><url>%s</url><stream>%s</stream><ticket>%s</ticket><bitrate>%s</bitrate></rtmp></getpublishstatus>`,
			publishStatusv1.Token,
			publishStatusv1.URL,
			publishStatusv1.Stream,
			publishStatusv1.Ticket,
			publishStatusv1.Bitrate)

		publishStatus2 = `<?xml version="1.0" encoding="utf-8"?>
<getpublishstatus status="fail" time="1486052405"><error><code>notfound</code></error></getpublishstatus>`
	)

	responce := publishStatus1
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, responce)
	}))
	defer ts.Close()

	ac := NewAccount("mail", "pass", "example")
	ps1, err := publishStatusImpl(ts.URL, ac)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(ps1, publishStatusv1) {
		t.Fatalf("Should be : %#v\nbut : %#v", publishStatusv1, ps1)
	}

	responce = publishStatus2
	_, err = publishStatusImpl(ts.URL, &Account{Usersession: "example"})
	if err == nil {
		t.Fatal("should be fail")
	}
}
