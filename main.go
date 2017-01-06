package main

// Application holds app settings and valuables
type Application struct {
	SavePath      string
	SettingsSlots SettingsSlots
}

// Application global information
var (
	AppName = "Nagome"
	Version string
	App     Application
)

func main() {
	RunCli()
	return
}
