package accounts

import (
	"log"
	"net/http"
	"time"

	"github.com/adnaan/authn"
	pwc "github.com/adnaan/fir/controller"
)

type SettingsView struct {
	pwc.DefaultView
	Auth *authn.API
}

func (s *SettingsView) Content() string {
	return "./templates/views/accounts/settings"
}

func (s *SettingsView) Layout() string {
	return "./templates/layouts/app.html"
}

func (s *SettingsView) OnEvent(st pwc.Socket) error {
	st.Store("settings").UpdateProp("profile_loading", true)
	defer func() {
		time.Sleep(1 * time.Second)
		st.Store("settings").UpdateProp("profile_loading", true)
	}()
	switch st.Event().ID {
	case "account/update":
		return s.UpdateProfile(st)
	case "account/delete":
		return s.DeleteAccount(st)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", st.Event())
	}
	return nil
}

func (s *SettingsView) OnRequest(w http.ResponseWriter, r *http.Request) (pwc.Status, pwc.Data) {
	if r.Method != "GET" {
		return pwc.Status{Code: 405}, nil
	}
	userID, _ := r.Context().Value(authn.AccountIDKey).(string)
	acc, err := s.Auth.GetAccount(r.Context(), userID)
	if err != nil {
		return pwc.Status{Code: 200}, nil
	}

	name := ""
	m := acc.Attributes().Map()
	if m != nil {
		name, _ = m.String("name")
	}

	return pwc.Status{Code: 200}, pwc.Data{
		"is_logged_in": true,
		"email":        acc.Email(),
		"name":         name,
	}
}

func (s *SettingsView) UpdateProfile(st pwc.Socket) error {
	req := new(ProfileRequest)
	if err := st.Event().DecodeParams(req); err != nil {
		return err
	}
	rCtx := st.Request().Context()
	userID, _ := rCtx.Value(authn.AccountIDKey).(string)
	acc, err := s.Auth.GetAccount(rCtx, userID)
	if err != nil {
		return err
	}
	if err := acc.Attributes().Set(rCtx, "name", req.Name); err != nil {
		return err
	}
	if req.Email != "" && req.Email != acc.Email() {
		if err := acc.ChangeEmail(rCtx, req.Email); err != nil {
			return err
		}
		st.Store("settings").UpdateProp("change_email", true)
	}

	st.Morph("#account_form", "account_form", pwc.Data{
		"name":  req.Name,
		"email": acc.Email(),
	})

	return nil
}

func (s *SettingsView) DeleteAccount(st pwc.Socket) error {
	rCtx := st.Request().Context()
	userID, _ := rCtx.Value(authn.AccountIDKey).(string)
	acc, err := s.Auth.GetAccount(rCtx, userID)
	if err != nil {
		return err
	}
	if err := acc.Delete(rCtx); err != nil {
		return err
	}
	st.Reload()
	return nil
}
