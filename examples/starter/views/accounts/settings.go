package accounts

import (
	"context"
	"log"
	"net/http"

	"github.com/adnaan/authn"
	"github.com/adnaan/fir"
)

type SettingsView struct {
	fir.DefaultView
	Auth *authn.API
}

func (s *SettingsView) Content() string {
	return "./templates/views/accounts/settings"
}

func (s *SettingsView) Layout() string {
	return "./templates/layouts/app.html"
}

func (s *SettingsView) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "account-update":
		return s.UpdateProfile(event)
	case "account-delete":
		return s.DeleteAccount(event)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	return nil
}

func (s *SettingsView) OnGet(w http.ResponseWriter, r *http.Request) fir.Pagedata {
	userID, _ := r.Context().Value(authn.AccountIDKey).(string)
	acc, err := s.Auth.GetAccount(r.Context(), userID)
	if err != nil {
		return fir.ErrBadRequest(err)
	}

	name := ""
	m := acc.Attributes().Map()
	if m != nil {
		name, _ = m.String("name")
	}

	return fir.Pagedata{
		Data: map[string]any{
			"is_logged_in": true,
			"email":        acc.Email(),
			"name":         name,
		}}
}

func (s *SettingsView) UpdateProfile(event fir.Event) fir.Patchset {
	req := new(ProfileRequest)
	if err := event.DecodeFormParams(req); err != nil {
		return fir.PatchError(err)
	}
	rCtx := event.RequestContext()
	userID, _ := rCtx.Value(authn.AccountIDKey).(string)
	acc, err := s.Auth.GetAccount(rCtx, userID)
	if err != nil {
		return fir.PatchError(err)
	}
	if err := acc.Attributes().Set(rCtx, "name", req.Name); err != nil {
		return fir.PatchError(err)
	}
	var patchset fir.Patchset
	if req.Email != "" && req.Email != acc.Email() {
		if err := acc.ChangeEmail(rCtx, req.Email); err != nil {
			return fir.PatchError(err)
		}
		patchset = append(patchset, fir.Store{
			Name: "settings",
			Data: map[string]any{
				"change_email": true,
			},
		})
	}

	patchset = append(patchset, fir.Morph{
		Selector: "#account_form",
		HTML: &fir.Render{
			Template: "account_form",
			Data: map[string]any{
				"name":  req.Name,
				"email": acc.Email(),
			},
		},
	})

	return patchset
}

func (s *SettingsView) DeleteAccount(event fir.Event) fir.Patchset {
	rCtx := event.RequestContext()
	userID, _ := rCtx.Value(authn.AccountIDKey).(string)
	acc, err := s.Auth.GetAccount(context.Background(), userID)
	if err != nil {
		return fir.PatchError(err)
	}
	if err := acc.Delete(rCtx); err != nil {
		return fir.PatchError(err)
	}
	return fir.Patchset{fir.Reload{}}
}
