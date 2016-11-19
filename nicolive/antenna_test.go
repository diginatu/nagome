package nicolive

import (
	"fmt"
	"strings"
	"testing"
)

func TestAntennaLoginParseProc(t *testing.T) {
	const testTicket = "nicolive_antenna_8888888888888888888888888888"
	resp := strings.NewReader(`<?xml version="1.0" encoding="utf-8"?>
		<nicovideo_user_response status="ok">
			<ticket>` + testTicket + `</ticket>
		</nicovideo_user_response>`)
	resperr := strings.NewReader(`<?xml version="1.0" encoding="utf-8"?>
		<nicovideo_user_response status="fail">
			<error>
				<code>1</code>
				<description>メールアドレスまたはパスワードが間違っているため、ログインできません</description>
			</error>
		</nicovideo_user_response>`)

	a := &Antenna{ac: &Account{Mail: "mail", Pass: "pass"}}

	nerr := a.loginParseProc(resperr)
	if nerr == nil {
		t.Fatal("Should be error")
	}

	nerr = a.loginParseProc(resp)
	if nerr != nil {
		t.Fatal(nerr)
	}
	if a.ticket != testTicket {
		t.Fatalf("Should be %v but %v", testTicket, a.ticket)
	}
}

func TestAntennaAdminParseProc(t *testing.T) {
	const (
		testAddr   = "test.addr"
		testPort   = "2525"
		testThread = "1000000000"
		testComms1 = "co1"
		testComms2 = "ch1234567"
		testComms3 = "co2345471"
	)

	resp := strings.NewReader(`<?xml version="1.0" encoding="utf-8"?>
		<getalertstatus status="ok" time="1477128042">
			<user_id>123</user_id>
			<user_hash>aaaaaaaaa1AAAAAAA2-bbb3BBB4</user_hash>
			<user_name>ユーザ名</user_name>
			<user_prefecture>1</user_prefecture>
			<user_age>30</user_age>
			<user_sex>1</user_sex>
			<communities>
				<community_id>` + testComms1 + `</community_id>
				<community_id>` + testComms2 + `</community_id>
				<community_id>` + testComms3 + `</community_id>
			</communities>
			<ms>
				<addr>` + testAddr + `</addr>
				<port>` + testPort + `</port>
				<thread>` + testThread + `</thread>
			</ms>
		</getalertstatus>`)

	resperr := strings.NewReader(`<?xml version="1.0" encoding="utf-8"?>
		<getalertstatus status="fail" time="1477129105">
			<error><code>incorrect_account_data</code></error>
		</getalertstatus>`)

	a := &Antenna{ac: &Account{Mail: "mail", Pass: "pass"}}

	nerr := a.adminParseProc(resperr)
	if nerr == nil {
		t.Fatal("Should be error")
	}

	nerr = a.adminParseProc(resp)
	if nerr != nil {
		t.Fatal(nerr)
	}
	t.Log(a)
	if a.addr != testAddr {
		t.Fatalf("Should be %v but %v", a.addr, testAddr)
	}
	if a.port != testPort {
		t.Fatalf("Should be %v but %v", a.port, testPort)
	}
	if a.thread != testThread {
		t.Fatalf("Should be %v but %v", a.thread, testThread)
	}
	if a.Following[0] != testComms1 ||
		a.Following[1] != testComms2 ||
		a.Following[2] != testComms3 {
		t.Fatalf("Should be [%v %v %v] but %v", testComms1, testComms2, testComms3, a.Following)
	}
}

type testEvProc struct {
	E *Event
}

func (t *testEvProc) ProceedNicoEvent(e *Event) {
	t.E = e
}
func TestAntennaProcceedMessage(t *testing.T) {
	var (
		testMes1    = `<thread resultcode="0" thread="1000000000" last_res="25286040" ticket="0x1a2b3c" revision="1" server_time="1477577699"/>`
		testBroadID = "123456789"
		testCommID  = "co123456"
		testUserID  = "98765432"
		testMes2    = fmt.Sprintf(`<chat thread="1000000000" no="25286040" date="1477577698" date_usec="711918" user_id="394" premium="2">%s,%s,%s</chat>`, testBroadID, testCommID, testUserID)
	)

	ev := &testEvProc{}
	a := &Antenna{connection: &connection{Ev: ev}}

	a.proceedMessage(testMes1)
	if ev.E.Type != EventTypeAntennaOpen {
		t.Fatalf("Should be %v but %v", ev.E.Type, EventTypeAntennaOpen)
	}
	a.proceedMessage(testMes2)
	if ev.E.Type != EventTypeAntennaGot {
		t.Fatalf("Should be %v but %v", ev.E.Type, EventTypeAntennaGot)
	}
	ai, ok := ev.E.Content.(AntennaItem)
	if !ok {
		t.Fatalf("Should be AntennaItem but %v", ev.E)
	}
	if ai.BroadID != testBroadID || ai.CommunityID != testCommID || ai.UserID != testUserID {
		t.Fatalf("Should be {%s,%s,%s} but %v", testBroadID, testCommID, testUserID, ai)
	}
}
