package accounts

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/adnaan/authn"
	fir "github.com/adnaan/fir/controller"
)

type ConfirmMagicView struct {
	fir.DefaultView
	Auth *authn.API
}

func (c *ConfirmMagicView) Content() string {
	return "./templates/views/accounts/confirm_magic"
}

func (c *ConfirmMagicView) Layout() string {
	return "./templates/layouts/index.html"
}

func (c *ConfirmMagicView) OnRequest(w http.ResponseWriter, r *http.Request) (fir.Status, fir.Data) {
	if r.Method != "GET" {
		return fir.Status{Code: 405}, nil
	}
	token := chi.URLParam(r, "token")
	err := c.Auth.LoginWithPasswordlessToken(w, r, token)
	if err != nil {
		return fir.Status{Code: 200}, nil
	}
	redirectTo := "/app"
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	return fir.Status{Code: 200}, nil
}
