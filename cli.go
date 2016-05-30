package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/diginatu/nagome/nicolive"
)

var (
	// Logger is logger in this app
	Logger        *log.Logger
	printVersion  bool
	printHelp     bool
	debugToStderr bool
	standAlone    bool
)

func init() {
	// set command line options
	flag.StringVar(&App.SavePath, "savepath",
		findUserConfigPath(), "Set <directory> to save directory.")
	flag.BoolVar(&printHelp, "h", false, "Print this help.")
	flag.BoolVar(&printHelp, "help", false, "Print this help.")
	flag.BoolVar(&printVersion, "version", false, "Print version information.")
	flag.BoolVar(&debugToStderr, "dbgtostd", false,
		`Output debug information to stderr.
	Without this option, output to the log file in the save directory.`)
	flag.BoolVar(&standAlone, "standalone", false, `Run in stand alone mode (CUI).`)
}

// RunCli processes flags and io
func RunCli() {
	flag.Parse()

	if printHelp {
		flag.Usage()
		os.Exit(0)
	}
	if printVersion {
		fmt.Println(App.Name, " ", App.Version)
		os.Exit(0)
	}

	if err := os.MkdirAll(App.SavePath, 0777); err != nil {
		log.Fatal("could not make save directory\n" + err.Error())
	}

	var file *os.File
	if debugToStderr {
		file = os.Stderr
	} else {
		var err error
		file, err = os.Create(filepath.Join(App.SavePath, "info.log"))
		if err != nil {
			log.Fatal("could not open log file\n" + err.Error())
		}
	}
	defer file.Close()
	Logger = log.New(file, "", log.Lshortfile|log.Ltime)

	if standAlone {
		standAloneMode()
		return
	}

	clientMode()
}

func standAloneMode() {
	stdinReader := bufio.NewReader(os.Stdin)

	var ac nicolive.Account
	ac.Load(filepath.Join(App.SavePath, "userData.yml"))

	var l nicolive.LiveWaku

	for {
		fmt.Println("input broad URL or ID (or empty to quit) :")
		brdtx, err := stdinReader.ReadString('\n')
		if err != nil || brdtx == "\n" {
			return
		}
		brdtx = brdtx[:len(brdtx)-1]

		brdRg := regexp.MustCompile("(lv|co)\\d+")
		broadMch := brdRg.FindString(brdtx)
		if broadMch == "" {
			fmt.Println("invalid text")
			continue
		}

		l = nicolive.LiveWaku{Account: &ac, BroadID: broadMch}
		break
	}

	nicoerr := l.FetchInformation()
	if nicoerr != nil {
		Logger.Fatalln(nicoerr)
	}

	commconn := nicolive.NewCommentConnection(&l, nil)
	commconn.Connect()

	for {
		text, err := stdinReader.ReadString('\n')
		if err != nil {
			commconn.Disconnect()
			return
		}
		text = text[:len(text)-1]

		switch text {
		case ":q":
			commconn.Disconnect()
			return
		default:
			commconn.SendComment(text, false)
		}
	}
}

func clientMode() {
	stdinReader := bufio.NewReader(os.Stdin)

	var ac nicolive.Account
	ac.Load(filepath.Join(App.SavePath, "userData.yml"))

	var l nicolive.LiveWaku
	var commconn *nicolive.CommentConnection

	dec := json.NewDecoder(stdinReader)
	for {
		var m Message
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			Logger.Println(err)
			continue
		}

		if m.Domain == "Nagome" {
			switch m.Func {
			case NagomeMess[FuncnBroadQuery].Funcn:
				switch m.Command {
				case NagomeMess[FuncnBroadQuery].Commands[CommBroadQueryConnect]:
					var cm BroadConnect
					if err := json.Unmarshal(m.Content, &cm); err != nil {
						Logger.Println("error:", err)
						continue
					}

					brdRg := regexp.MustCompile("(lv|co)\\d+")
					broadMch := brdRg.FindString(cm.BroadID)
					if broadMch == "" {
						Logger.Println("invalid text")
						continue
					}

					l = nicolive.LiveWaku{Account: &ac, BroadID: broadMch}

					nicoerr := l.FetchInformation()
					if nicoerr != nil {
						Logger.Println(nicoerr)
						continue
					}

					commconn = nicolive.NewCommentConnection(&l, nil)
					nicoerr = commconn.Connect()
					if nicoerr != nil {
						Logger.Println(nicoerr)
						continue
					}

					defer commconn.Disconnect()
				default:
					Logger.Println("invalid Command in message")
				}
			case NagomeMess[FuncnAccountQuery].Funcn:
				switch m.Command {
				case NagomeMess[FuncnAccountQuery].Commands[CommAccountLogin]:
					err := ac.Login()
					if err != nil {
						Logger.Fatalln(err)
						continue
					}
					Logger.Println("logged in")
				case NagomeMess[FuncnAccountQuery].Commands[CommAccountSave]:
					ac.Save(filepath.Join(App.SavePath, "userData.yml"))
				case NagomeMess[FuncnAccountQuery].Commands[CommAccountLoad]:
					ac.Load(filepath.Join(App.SavePath, "userData.yml"))
				default:
					Logger.Println("invalid Command in message")
				}
			default:
				Logger.Println("invalid Func in message")
			}
		}

	}
}
