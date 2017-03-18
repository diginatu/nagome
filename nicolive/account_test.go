package nicolive

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestAccountStringer(t *testing.T) {
	a := &Account{
		Mail:        "test@example.com",
		Pass:        "1234",
		Usersession: "abcde",
	}
	expect := "Account{test@..}"
	got := fmt.Sprint(a)
	if got != expect {
		t.Fatalf("Should be %v but %v", expect, got)
	}

	a.Mail = "testaccount@example.com"
	expect = "Account{testa..}"
	got = fmt.Sprint(a)
	if got != expect {
		t.Fatalf("Should be %v but %v", expect, got)
	}

	a.Mail = "test"
	expect = "Account{test..}"
	got = fmt.Sprint(a)
	if got != expect {
		t.Fatalf("Should be %v but %v", expect, got)
	}
}

func TestAccountLoad(t *testing.T) {
	a, err := AccountLoad("testdata/success.yaml")
	if err != nil {
		t.Fatal(err)
	}

	expect := "test@example.com"
	if a.Mail != expect {
		t.Fatalf("Mail should be %v but %v", expect, a.Mail)
	}

	expect = "1234"
	if a.Pass != expect {
		t.Fatalf("Pass should be %v but %v", expect, a.Pass)
	}

	expect = "abcde"
	if a.Usersession != expect {
		t.Fatalf("Usersession should be %v but %v", expect, a.Pass)
	}
}
func TestAccountLoadFail(t *testing.T) {
	_, err := AccountLoad("testdata/failing.yaml")
	if err == nil {
		t.Fatal("Should be fail")
	}
}

func TestAccountSave(t *testing.T) {
	a := &Account{
		Mail:        "test@example.com",
		Pass:        "1234",
		Usersession: "abcde",
	}

	f, err := ioutil.TempFile("", "nagome")
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := os.Remove(f.Name())
		if err != nil {
			t.Fatal(err)
		}
	}()

	err = a.Save(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	b, err := AccountLoad(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if a.Mail != b.Mail {
		t.Fatalf("Should be %v but %v", a.Mail, b.Mail)
	}
	if a.Pass != b.Pass {
		t.Fatalf("Should be %v but %v", a.Pass, b.Pass)
	}
	if a.Usersession != b.Usersession {
		t.Fatalf("Should be %v but %v", a.Usersession, b.Usersession)
	}
}
