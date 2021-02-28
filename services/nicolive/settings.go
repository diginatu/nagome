package nicolive

// A Settings represents a settings of nicolive.
type Settings struct {
	UserNameGet  bool `yaml:"user_name_get"`
	OwnerComment bool `yaml:"owner_comment"`
}

// NewSettings creates new SettingsSlot with default values.
func NewSettings() *Settings {
	return &Settings{
		UserNameGet:  false,
		OwnerComment: true,
	}
}

// Duplicate creates new copy.
func (s *Settings) Duplicate() Settings {
	return *s
}
