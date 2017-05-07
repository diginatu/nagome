package viewer

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// A SettingsSlot represents a settings of Nagome.
type SettingsSlot struct {
	Name               string          `yaml:"name"                  json:"name"`
	AutoSaveTo0Slot    bool            `yaml:"auto_save_to0_slot"    json:"auto_save_to0_slot"`
	UserNameGet        bool            `yaml:"user_name_get"         json:"user_name_get"`
	AutoFollowNextWaku bool            `yaml:"auto_follow_next_waku" json:"auto_follow_next_waku"`
	OwnerComment       bool            `yaml:"owner_comment"         json:"owner_comment"`
	PluginDisable      map[string]bool `yaml:"plugin_disable"        json:"plugin_disable"`
}

// NewSettingsSlot creates new SettingsSlot with default values.
func NewSettingsSlot() *SettingsSlot {
	return &SettingsSlot{
		Name:               "New Settings",
		AutoSaveTo0Slot:    true,
		UserNameGet:        false,
		AutoFollowNextWaku: true,
		OwnerComment:       true,
		PluginDisable:      make(map[string]bool),
	}
}

// UnmarshalYAML is a function for implementing yaml.Unmarshaler.
func (s *SettingsSlot) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Using same struct type causes recursive function call.
	type UnmarshalSettingsSlot SettingsSlot
	ns := (*UnmarshalSettingsSlot)(NewSettingsSlot())
	err := unmarshal(ns)
	if err != nil {
		return err
	}
	// Copy as values not rewriting the pointer.
	*s = *(*SettingsSlot)(ns)
	return nil
}

// Duplicate creates new copy.
func (s *SettingsSlot) Duplicate() SettingsSlot {
	var ns SettingsSlot
	ns = *s
	ns.PluginDisable = make(map[string]bool)
	for k, c := range s.PluginDisable {
		ns.PluginDisable[k] = c
	}
	return ns
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
func (ss *SettingsSlots) Save(path string) error {
	s, err := yaml.Marshal(ss)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, s, 0600)
}

// Load loads from a file.
func (ss *SettingsSlots) Load(path string) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(f, ss)
}
