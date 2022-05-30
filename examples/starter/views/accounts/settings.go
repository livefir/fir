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

func (s *SettingsView) OnLiveEvent(ctx pwc.Context) error {
	ctx.Store("settings").UpdateProp("profile_loading", true)
	defer func() {
		time.Sleep(1 * time.Second)
		ctx.Store("settings").UpdateProp("profile_loading", true)
	}()
	switch ctx.Event().ID {
	case "account/update":
		return s.UpdateProfile(ctx)
	case "account/delete":
		return s.DeleteAccount(ctx)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", ctx.Event())
	}
	return nil
}

func (s *SettingsView) OnMount(w http.ResponseWriter, r *http.Request) (pwc.Status, pwc.M) {
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

	return pwc.Status{Code: 200}, pwc.M{
		"is_logged_in": true,
		"email":        acc.Email(),
		"name":         name,
	}
}

func (s *SettingsView) UpdateProfile(ctx pwc.Context) error {
	req := new(ProfileRequest)
	if err := ctx.Event().DecodeParams(req); err != nil {
		return err
	}
	rCtx := ctx.Request().Context()
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
		ctx.Store("settings").UpdateProp("change_email", true)
	}

	ctx.Morph("#account_form", "account_form", pwc.M{
		"name":  req.Name,
		"email": acc.Email(),
	})

	return nil
}

func (s *SettingsView) DeleteAccount(ctx pwc.Context) error {
	rCtx := ctx.Request().Context()
	userID, _ := rCtx.Value(authn.AccountIDKey).(string)
	acc, err := s.Auth.GetAccount(rCtx, userID)
	if err != nil {
		return err
	}
	if err := acc.Delete(rCtx); err != nil {
		return err
	}
	ctx.Reload()
	return nil
}
