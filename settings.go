package main

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// A SettingsSlot represents a settings of Nagome.
type SettingsSlot struct {
	SettingsName       string          `yaml:"settings_name"         json:"settings_name"`
	AutoSaveTo0Slot    bool            `yaml:"auto_save_to0_slot"    json:"auto_save_to0_slot"`
	UserNameGet        bool            `yaml:"user_name_get"         json:"user_name_get"`
	AutoFollowNextWaku bool            `yaml:"auto_follow_next_waku" json:"auto_follow_next_waku"`
	PluginDisable      map[string]bool `yaml:"plugin_disable"        json:"plugin_disable"`
}

// NewSettingsSlot creates new SettingsSlot with default values.
func NewSettingsSlot() *SettingsSlot {
	return &SettingsSlot{
		SettingsName:       "New Settings",
		AutoSaveTo0Slot:    true,
		UserNameGet:        false,
		AutoFollowNextWaku: true,
		PluginDisable:      make(map[string]bool),
	}
}

// SettingsSlots is struct for multiple configs file.
type SettingsSlots struct {
	Config []*SettingsSlot `yaml:"config" json:"config"`
}

// Add adds given slot to the list.
func (ss *SettingsSlots) Add(s *SettingsSlot) {
	ss.Config = append(ss.Config, s)
}

// Save saves to a file.
func (ss *SettingsSlots) Save() error {
	s, err := yaml.Marshal(ss)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(App.SavePath, settingsFileName), s, 0600)
	if err != nil {
		return err
	}

	return nil
}

// Load loads from a file.
func (ss *SettingsSlots) Load() error {
	f, err := ioutil.ReadFile(filepath.Join(App.SavePath, settingsFileName))
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(f, ss)
	if err != nil {
		return err
	}

	return nil
}
