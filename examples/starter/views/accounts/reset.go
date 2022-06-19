package accounts

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/adnaan/authn"
	"github.com/adnaan/fir"
)

type ResetView struct {
	fir.DefaultView
	Auth *authn.API
}

func (rv *ResetView) Content() string {
	return "./templates/views/accounts/reset"
}

func (rv *ResetView) Layout() string {
	return "./templates/layouts/index.html"
}

func (rv *ResetView) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "account/reset":
		r := new(ResetReq)
		if err := event.DecodeParams(r); err != nil {
			return nil
		}
		if r.ConfirmPassword != r.Password {
			return nil
			// return nil, fmt.Errorf("%w", errors.New("passwords don't match"))
		}
		if err := rv.Auth.ConfirmRecovery(event.RequestContext(), r.Token, r.Password); err != nil {
			return nil
		}
		return fir.Patchset{fir.Store{
			Name: "reset",
			Data: map[string]any{"password_reset": true},
		}}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	return nil
}

func (rv *ResetView) OnRequest(w http.ResponseWriter, r *http.Request) (fir.Status, fir.Data) {
	token := chi.URLParam(r, "token")
	return fir.Status{Code: 200}, fir.Data{
		"token": token,
	}
}

type ResetReq struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Token           string `json:"token"`
}
