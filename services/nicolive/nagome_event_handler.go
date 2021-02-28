package nicolive

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

	"github.com/diginatu/nagome/api"
	"github.com/diginatu/nagome/services/utils"
)

// Constant values for processing plugin messages.
const (
	NumFetchInformaionRetry = 3
	accountFileName         = "account.yml"
)

var (
	broadIDRegex = regexp.MustCompile(`(lv|co)\d+`)
)

func ensureAccountLoaded(v *Viewer) {
	// Ensure account is loaded
	if v.Ac == nil {
		ac, err := AccountLoad(filepath.Join(v.savePath, accountFileName))
		if err != nil {
			v.log.Println(err)
			v.Ac = new(Account)
			v.Evch <- api.NewMessageMust(api.DomainUI, api.CommUIConfigAccount, nil)
		} else {
			v.Ac = ac
		}
	}
}

func NagomeQueryBroadConnectHandler(v *Viewer, m *api.Message) error {
	ensureAccountLoaded(v)

	var err error
	var ct api.CtQueryBroadConnect
	if err = json.Unmarshal(m.Content, &ct); err != nil {
		return MakeError(ErrOther, "JSON error in the content : "+err.Error())
	}

	broadMch := broadIDRegex.FindString(ct.BroadID)
	if broadMch == "" {
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "invalid BroadID", "no valid BroadID found in the ID text", v.Evch, v.log)
		return MakeError(ErrOther, "no valid BroadID found in the ID text")
	}

	lw := &LiveWaku{Account: v.Ac, BroadID: broadMch}

	err = lw.FetchInformation()
	if err != nil {
		nerr, ok := err.(Error)
		if ok {
			if nerr.Type() == ErrNetwork {
				ct.RetryN++
				v.log.Printf("Failed to connect to %s.\n", ct.BroadID)
				v.log.Printf("FetchInformation : %s\n", nerr.Error())
				if ct.RetryN >= NumFetchInformaionRetry {
					v.log.Println("Reached the limit of retrying.")
					return nerr
				}
				v.log.Println("Retrying...")
				go func() {
					<-time.After(time.Second)
					v.Evch <- api.NewMessageMust(api.DomainQuery, api.CommQueryBroadConnect, ct)
				}()
			}
		}
		return err
	}

	v.Disconnect()

	v.Lw = lw
	v.Cmm, err = CommentConnect(context.TODO(), *v.Lw, v)
	if err != nil {
		return err
	}

	if lw.IsUserOwner() {
		ps, err := PublishStatus(lw.BroadID, v.Ac)
		if err != nil {
			return err
		}
		v.Lw.OwnerCommentToken = ps.Token
	}

	return nil
}

func NagomeQueryBroadDisconnectHandler(v *Viewer) error {
	return v.Disconnect()
}

func NagomeQueryBroadSendCommentHandler(v *Viewer, m *api.Message) error {
	if v.Cmm == nil {
		return MakeError(ErrSendComment, "not connected to live")
	}
	if v.Lw == nil {
		return MakeError(ErrSendComment, "v.Lw is nil")
	}

	var ct api.CtQueryBroadSendComment
	if err := json.Unmarshal(m.Content, &ct); err != nil {
		return MakeError(ErrOther, "JSON error in the content : "+err.Error())
	}

	isowner := false
	if v.Lw.IsUserOwner() {
		isowner = v.settings.OwnerComment
		if ct.Type != "" {
			isowner = ct.Type == api.CtQueryBroadSendCommentTypeOwner
		}
	}

	if isowner {
		rq := CommentOwnerRequest{
			Text:        ct.Text,
			Color:       "",
			IsPermanent: false,
			UserName:    "",
		}
		err := CommentOwner(v.Lw.BroadID, http.MethodPut, &rq, v.Ac)
		if err != nil {
			return err
		}
	} else {
		v.Cmm.SendComment(ct.Text, ct.Iyayo)
	}

	return nil
}

func NagomeQueryAccountSetHandler(v *Viewer, m *api.Message) error {
	if v.Ac == nil {
		return MakeError(ErrOther, "Account data (v.Ac) is nil.")
	}
	var ct api.CtQueryAccountSet
	if err := json.Unmarshal(m.Content, &ct); err != nil {
		return MakeError(ErrOther, "JSON error in the content : "+err.Error())
	}
	if ct.Mail != "" {
		v.Ac.Mail = ct.Mail
	}
	if ct.Pass != "" {
		v.Ac.Pass = ct.Pass
	}
	if ct.Usersession != "" {
		v.Ac.Usersession = ct.Usersession
	}

	return nil
}

func NagomeQueryAccountLoginHandler(v *Viewer, m *api.Message) error {
	ensureAccountLoaded(v)

	err := v.Ac.Login()
	if err != nil {
		if nerr, ok := err.(Error); ok {
			utils.EmitEvNewNotification(
				api.CtUINotificationTypeWarn, "login error", nerr.Description(),
				v.Evch, v.log)
		} else {
			utils.EmitEvNewNotification(
				api.CtUINotificationTypeWarn, "login error", err.Error(),
				v.Evch, v.log)
		}
		return err
	}
	v.log.Println("logged in")
	utils.EmitEvNewNotification(
		api.CtUINotificationTypeInfo, "login succeeded", "login succeeded",
		v.Evch, v.log)

	return nil
}

func NagomeQueryAccountLoadHandler(v *Viewer, m *api.Message) error {
	var err error
	v.Ac, err = AccountLoad(filepath.Join(v.savePath, accountFileName))
	return err
}

func NagomeQueryAccountSaveHandler(v *Viewer, m *api.Message) error {
	ensureAccountLoaded(v)

	return v.Ac.Save(filepath.Join(v.savePath, accountFileName))
}

func NagomeQueryUserSetHandler(v *Viewer, m *api.Message) error {
	var ct User // CtQueryUserSet
	if err := json.Unmarshal(m.Content, &ct); err != nil {
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Storing the user info failed", "JSON parse error", v.Evch, v.log)
		return MakeError(ErrOther, "JSON error in the content : "+err.Error())
	}

	err := v.userDB.Store(&ct)
	if err != nil {
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Storing the user info failed", "DB error : "+err.Error(), v.Evch, v.log)
		return err
	}

	v.Evch <- api.NewMessageMust(api.DomainNagome, api.CommNagomeUserUpdate, api.CtNagomeUserUpdate(ct))

	return nil
}

func NagomeQueryUserSetNameHandler(v *Viewer, m *api.Message) error {
	var ct api.CtQueryUserSetName
	if err := json.Unmarshal(m.Content, &ct); err != nil {
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Storing the user name failed", "JSON parse error", v.Evch, v.log)
		return MakeError(ErrOther, "JSON error in the content : "+err.Error())
	}

	if ct.Name == "" {
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Blank name", "You can't set blank name", v.Evch, v.log)
		return MakeError(ErrOther, "format error : Name is empty")
	}

	user, err := v.userDB.Fetch(ct.ID)
	if err != nil {
		nerr, ok := err.(Error)
		if ok && nerr.Type() == ErrDBUserNotFound {
			user, err = v.CheckIntervalAndCreateUser(ct.ID)
			if err != nil {
				utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Storing the user name failed", "DB error", v.Evch, v.log)
				return MakeError(ErrOther, "Storing the user name failed: DB error : "+err.Error())
			}
		} else {
			utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Storing the user name failed", "DB error", v.Evch, v.log)
			return MakeError(ErrOther, "Storing the user name failed: DB error : "+err.Error())
		}
	}

	user.Name = ct.Name
	err = v.userDB.Store(user)
	if err != nil {
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Storing the user name failed", "DB error : "+err.Error(), v.Evch, v.log)
		return err
	}

	v.Evch <- api.NewMessageMust(api.DomainNagome, api.CommNagomeUserUpdate, api.CtNagomeUserUpdate(*user))
	return nil
}

func NagomeQueryUserDeleteHandler(v *Viewer, m *api.Message) error {
	var ct api.CtQueryUserDelete
	if err := json.Unmarshal(m.Content, &ct); err != nil {
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Deleting the user info failed", "JSON parse error", v.Evch, v.log)
		return MakeError(ErrOther, "JSON error in the content : "+err.Error())
	}

	err := v.userDB.Remove(ct.ID)
	if err != nil {
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Removing the user info failed", "DB error : "+err.Error(), v.Evch, v.log)
		return err
	}

	usr := api.CtNagomeUserUpdate{
		ID:           ct.ID,
		Name:         "",
		CreateTime:   time.Unix(0, 0),
		Is184:        Is184UserID(ct.ID),
		ThumbnailURL: "",
	}
	v.Evch <- api.NewMessageMust(api.DomainNagome, api.CommNagomeUserUpdate, usr)
	return nil
}

func NagomeQueryUserFetchHandler(v *Viewer, m *api.Message) error {
	var ct api.CtQueryUserFetch
	if err := json.Unmarshal(m.Content, &ct); err != nil {
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Fetching the user info failed", "JSON parse error", v.Evch, v.log)
		return MakeError(ErrOther, "JSON error in the content : "+err.Error())
	}

	user, err := v.CheckIntervalAndCreateUser(ct.ID)
	if err != nil {
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Removing the user info failed", "DB error : "+err.Error(), v.Evch, v.log)
		return err
	}

	userCurrent, err := v.userDB.Fetch(ct.ID)
	if err != nil {
		nerr, ok := err.(Error)
		if ok && nerr.Type() == ErrDBUserNotFound {
			// If the user was not in the DB
			userCurrent = user
		} else {
			utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Removing the user info failed", "DB error : "+err.Error(), v.Evch, v.log)
			return err
		}
	} else {
		userCurrent.Name = user.Name
		userCurrent.ThumbnailURL = user.ThumbnailURL
	}

	err = v.userDB.Store(userCurrent)
	if err != nil {
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Removing the user info failed", "DB error : "+err.Error(), v.Evch, v.log)
		return err
	}

	v.Evch <- api.NewMessageMust(api.DomainNagome, api.CommNagomeUserUpdate, api.CtNagomeUserUpdate(*userCurrent))

	return nil
}

func NagomeDirectUserGetHandler(v *Viewer, m *api.Message) error {
	var ct api.CtDirectUserGet
	if err := json.Unmarshal(m.Content, &ct); err != nil {
		return MakeError(ErrOther, "JSON error in the content : "+err.Error())
	}

	user, err := v.userDB.Fetch(ct.ID)
	if err != nil {
		utils.EmitEvNewNotification(api.CtUINotificationTypeWarn, "Removing the user info failed", "DB error : "+err.Error(), v.Evch, v.log)
		v.log.Printf("Removing the user info failed.\n DB error : %s", err.Error())
		return err
	}

	v.Evch <- api.NewMessageMust(api.DomainDirectngm, api.CommDirectngmUserGet, api.CtDirectngmUserGet(*user))

	return nil
}
