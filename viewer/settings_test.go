package viewer

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSettingsSlotsLoad(t *testing.T) {
	slotfilepath := filepath.Join(os.TempDir(), settingsFileName)

	ss := SettingsSlots{}
	testSetting := NewSettingsSlot()
	testSetting2 := NewSettingsSlot()
	testSetting2.UserNameGet = true
	testSetting2.PluginDisable["test"] = true
	ss.Add(testSetting)
	ss.Add(testSetting2)

	err := ss.Save(slotfilepath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.Remove(slotfilepath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	nss := SettingsSlots{}
	err = nss.Load(slotfilepath)
	if err != nil {
		t.Fatal(err)
	}

	if len(nss.Config) != len(ss.Config) {
		t.Fatalf("Setting length should be %v but %v", len(ss.Config), len(nss.Config))
	}
	for k, v := range ss.Config {
		t.Logf("%#v", nss.Config[k])
		if !reflect.DeepEqual(nss.Config[k], v) {
			t.Fatalf("Config[%d] Should be %v but %v", k, v, nss.Config[k])
		}
	}
}

func TestSettingsSlotsOldLoad(t *testing.T) {
	defaultss := NewSettingsSlot()

	ss := SettingsSlots{}
	err := ss.Load(filepath.Join("testdata", "old-setting", settingsFileName))
	if err != nil {
		t.Fatal(err)
	}

	// Settings that are set in the file
	if got := ss.Config[0].Name; got != "Old Settings" {
		t.Fatalf("Should be \"%v\" but \"%v\"", "Old Settings", got)
	}
	if got := ss.Config[0].AutoFollowNextWaku; got != true { // nolint: megacheck
		t.Fatalf("Should be %v but %v", true, got)
	}
	// Settings that are NOT set in the file
	if got := ss.Config[0].AutoSaveTo0Slot; got != defaultss.AutoSaveTo0Slot {
		t.Fatalf("Should be %v but %v", defaultss.AutoSaveTo0Slot, got)
	}
	if got := ss.Config[0].UserNameGet; got != defaultss.UserNameGet {
		t.Fatalf("Should be %v but %v", defaultss.UserNameGet, got)
	}
	if ss.Config[0].PluginDisable == nil {
		t.Fatalf("Should be initialized")
	}
}

func TestSettingsSlotDuplicate(t *testing.T) {
	s1 := *NewSettingsSlot()
	s1.PluginDisable["test"] = true
	s2 := s1.Duplicate()

	if !reflect.DeepEqual(s1, s2) {
		t.Fatalf("Should be %v but %v", s2, s1)
	}

	const onlyS2Key = "kepe"
	s2.PluginDisable[onlyS2Key] = true

	if s1.PluginDisable[onlyS2Key] {
		t.Fatalf("Should not share values")
	}
}
