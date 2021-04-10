package viewer

import (
	"io/ioutil"

	"github.com/diginatu/nagome/api"
	"github.com/diginatu/nagome/services/nicolive"
	"gopkg.in/yaml.v2"
)

// A SettingsSlot represents a settings of Nagome.
type SettingsSlot struct {
	Name            string          `yaml:"name"`
	AutoSaveTo0Slot bool            `yaml:"auto_save_to0_slot"`
	PluginDisable   map[string]bool `yaml:"plugin_disable"`

	Nicolive nicolive.Settings `yaml:"nicolive"`
}

// NewSettingsSlot creates new SettingsSlot with default values.
func NewSettingsSlot() *SettingsSlot {
	return &SettingsSlot{
		Name:            "New Settings",
		AutoSaveTo0Slot: true,
		PluginDisable:   make(map[string]bool),
	}
}

func NewSettingsSlotFromAPI(s *api.SettingsSlot) *SettingsSlot {
	return &SettingsSlot{
		Name:            s.Name,
		AutoSaveTo0Slot: s.AutoSaveTo0Slot,
		PluginDisable:   s.PluginDisable,
		Nicolive: nicolive.Settings{
			UserNameGet:  s.Nicolive.UserNameGet,
			OwnerComment: s.Nicolive.OwnerComment,
		},
	}
}

// API returns API representation of SettingsSlot
func (s *SettingsSlot) API() *api.SettingsSlot {
	return &api.SettingsSlot{
		Name:            s.Name,
		AutoSaveTo0Slot: s.AutoSaveTo0Slot,
		PluginDisable:   s.PluginDisable,

		Nicolive: api.SettingsNicolive{
			UserNameGet:  s.Nicolive.UserNameGet,
			OwnerComment: s.Nicolive.OwnerComment,
		},
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
	ns := *s
	ns.PluginDisable = make(map[string]bool)
	for k, c := range s.PluginDisable {
		ns.PluginDisable[k] = c
	}
	return ns
}

// SettingsSlots is struct for multiple configs file.
type SettingsSlots struct {
	Config []*SettingsSlot `yaml:"config"`
}

func NewSettingsSlotsFromAPI(ss *api.SettingsSlots) *SettingsSlots {
	ns := &SettingsSlots{
		Config: make([]*SettingsSlot, len(ss.Config)),
	}
	for i, s := range ss.Config {
		ns.Config[i] = NewSettingsSlotFromAPI(s)
	}

	return ns
}

// API returns API representation of SettingsSlots
func (ss *SettingsSlots) API() *api.SettingsSlots {
	slots := make([]*api.SettingsSlot, len(ss.Config))

	for i, s := range ss.Config {
		slots[i] = s.API()
	}

	return &api.SettingsSlots{
		Config: slots,
	}
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
