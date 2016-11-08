package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsSlots(t *testing.T) {
	setLogForTest()
	App.SavePath = os.TempDir()

	ss := SettingsSlots{}
	testSetting := &SettingsSlot{
		UserNameGet: false,
	}
	testSetting2 := &SettingsSlot{
		UserNameGet: true,
	}
	ss.Add(testSetting)
	ss.Add(testSetting2)

	err := ss.Save()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath.Join(App.SavePath, settingsFileName))

	err = ss.Load()
	if err != nil {
		t.Fatal(err)
	}

	if got := len(ss.Config); got != 2 {
		t.Fatalf("Should be %v but %v", 2, got)
	}
	if got := ss.Config[0]; !got.Equal(testSetting) {
		t.Fatalf("Should be %v but %v", testSetting, got)
	}
	if got := ss.Config[1]; !got.Equal(testSetting2) {
		t.Fatalf("Should be %v but %v", testSetting2, got)
	}
}
