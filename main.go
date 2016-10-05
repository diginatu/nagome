package main

// Application holds app settings and valuables
type Application struct {
	// Name is name of this app
	Name string
	// Version is version info
	Version string
	// SavePath is directory to hold save files
	SavePath string
}

var (
	// App is global Application settings and valuables for this app
	App = Application{Name: "Nagome", Version: "0.0.1"}
)

func main() {
	RunCli()
	return
}
