package main

// Application holds app settings and valuables
type Application struct {
	Name          string
	Version       string
	SavePath      string
	SettingsSlots SettingsSlots
}

var (
	// App is global Application settings and valuables for this app
	App = Application{Name: "Nagome", Version: "0.0.1"}
)

func main() {
	RunCli()
	return
}
