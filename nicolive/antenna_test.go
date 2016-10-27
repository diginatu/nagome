package nicolive

import "strings"
import "testing"

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

	a := NewAntenna(&Account{Mail: "mail", Pass: "pass"})

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
				<community_id>co2345471</community_id>
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

	a := NewAntenna(&Account{Mail: "mail", Pass: "pass"})

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

}
