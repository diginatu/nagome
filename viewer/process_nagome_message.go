package viewer

import (
	"encoding/json"

	"github.com/diginatu/nagome/api"
	"github.com/diginatu/nagome/services/nicolive"
)

func processNagomeMessage(cv *CommentViewer, m *api.Message) error {
	switch m.Domain {
	case api.DomainQuery:
		switch m.Command {
		case api.CommQueryBroadConnect:
			if err := nicolive.NagomeQueryBroadConnectHandler(cv.viewerNicolive, m); err != nil {
				return err
			}
			cv.cli.log.Println("connected")

		case api.CommQueryBroadDisconnect:
			if err := nicolive.NagomeQueryBroadDisconnectHandler(cv.viewerNicolive); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "Nicolive disconection failed: "+err.Error())
			}

		case api.CommQueryBroadSendComment:
			if err := nicolive.NagomeQueryBroadSendCommentHandler(cv.viewerNicolive, m); err != nil {
				return err
			}

		case api.CommQueryAccountSet:
			if err := nicolive.NagomeQueryAccountSetHandler(cv.viewerNicolive, m); err != nil {
				return err
			}

		case api.CommQueryAccountLogin:
			if err := nicolive.NagomeQueryAccountLoginHandler(cv.viewerNicolive, m); err != nil {
				return err
			}

		case api.CommQueryAccountLoad:
			if err := nicolive.NagomeQueryAccountLoadHandler(cv.viewerNicolive, m); err != nil {
				return err
			}

		case api.CommQueryAccountSave:
			if err := nicolive.NagomeQueryAccountSaveHandler(cv.viewerNicolive, m); err != nil {
				return err
			}

		case api.CommQueryLogPrint:
			var ct api.CtQueryLogPrint
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			cv.cli.log.Printf("plug[%s] %s\n", cv.PluginName(m.Plgno), ct.Text)

		case api.CommQuerySettingsSetCurrent:
			var ct api.CtQuerySettingsSetCurrent
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			cv.Settings = SettingsSlot{
				Name:            ct.Name,
				AutoSaveTo0Slot: ct.AutoSaveTo0Slot,
				PluginDisable:   ct.PluginDisable,
				Nicolive: nicolive.Settings{
					UserNameGet:  ct.Nicolive.UserNameGet,
					OwnerComment: ct.Nicolive.OwnerComment,
				},
			}

			for _, p := range cv.Pgns {
				p.SetState(!cv.Settings.PluginDisable[p.Name])
			}

		case api.CommQuerySettingsSetAll:
			var ct api.CtQuerySettingsSetAll
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			cv.cli.SettingsSlots = *NewSettingsSlotsFromAPI((*api.SettingsSlots)(&ct))

		case api.CommQueryPlugEnable:
			var ct api.CtQueryPlugEnable
			if err := json.Unmarshal(m.Content, &ct); err != nil {
				return nicolive.MakeError(nicolive.ErrOther, "JSON error in the content : "+err.Error())
			}

			pl, err := cv.Plugin(ct.No)
			if err != nil {
				return err
			}
			pl.SetState(ct.Enable)
			if cv.Settings.PluginDisable == nil {
				cv.Settings.PluginDisable = make(map[string]bool)
			}
			cv.Settings.PluginDisable[pl.Name] = !ct.Enable

		case api.CommQueryUserSet:
			if err := nicolive.NagomeQueryUserSetHandler(cv.viewerNicolive, m); err != nil {
				return err
			}

		case api.CommQueryUserSetName:
			if err := nicolive.NagomeQueryUserSetNameHandler(cv.viewerNicolive, m); err != nil {
				return err
			}

		case api.CommQueryUserDelete:
			if err := nicolive.NagomeQueryUserDeleteHandler(cv.viewerNicolive, m); err != nil {
				return err
			}

		case api.CommQueryUserFetch:
			if err := nicolive.NagomeQueryUserFetchHandler(cv.viewerNicolive, m); err != nil {
				return err
			}

		case api.CommDirectUserGet:
			if err := nicolive.NagomeDirectUserGetHandler(cv.viewerNicolive, m); err != nil {
				return err
			}

		default:
			return nicolive.MakeError(nicolive.ErrOther, "Message : invalid query command : "+m.Command)
		}
	}

	return nil
}

func processDirectMessage(cv *CommentViewer, m *api.Message) error {
	var t *api.Message
	var err error

	switch m.Command {
	case api.CommDirectngmAppVersion:
		t, err = api.NewMessage(api.DomainDirectngm, api.CommDirectngmAppVersion, api.CtDirectngmAppVersion{
			Name:    cv.cli.AppName,
			Version: cv.cli.Version,
		})
		if err != nil {
			return nicolive.ErrFromStdErr(err)
		}
	case api.CommDirectPlugList:
		c := api.CtDirectngmPlugList{Plugins: NewPluginsAPI(cv.Pgns)}
		t, err = api.NewMessage(api.DomainDirectngm, api.CommDirectngmPlugList, c)
		if err != nil {
			return nicolive.ErrFromStdErr(err)
		}
	case api.CommDirectSettingsCurrent:
		t, err = api.NewMessage(api.DomainDirectngm, api.CommDirectngmSettingsCurrent, cv.Settings.API())
		if err != nil {
			return nicolive.ErrFromStdErr(err)
		}
	case api.CommDirectSettingsAll:
		t, err = api.NewMessage(api.DomainDirectngm, api.CommDirectngmSettingsAll, cv.cli.SettingsSlots.API())
		if err != nil {
			return nicolive.ErrFromStdErr(err)
		}
	default:
		return nicolive.MakeError(nicolive.ErrOther, "Message : invalid query command : "+m.Command)
	}

	cv.Pgns[m.Plgno].WriteMess(t)
	return nil
}
