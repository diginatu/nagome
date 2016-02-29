package main

import (
	"path/filepath"

	"github.com/diginatu/nagome/nicolive"
)

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
	App = Application{Name: "Nagome", Version: "0.0"}
)

func main() {
	RunCli()

	var ac nicolive.Account
	ac.Load(filepath.Join(App.SavePath, "userData.yml"))
	//ac.Save(filepath.Join(App.SavePath, "userData.yml"))

	//err = ac.Login()
	//if err != nil {
	//Logger.Fatalln(err)
	//}

	l := nicolive.LiveWaku{Account: &ac, BroadID: "lv253955473"}
	nicoerr := l.FetchInformation()
	if nicoerr != nil {
		Logger.Fatalln(nicoerr)
	}

	//commentconn := nicolive.NewCommentConnection(&l)
	//commentconn.Connect()

	return
}
