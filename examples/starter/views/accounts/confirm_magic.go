package accounts

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/adnaan/authn"
	pwc "github.com/adnaan/fir/controller"
)

type ConfirmMagicView struct {
	pwc.DefaultView
	Auth *authn.API
}

func (c *ConfirmMagicView) Content() string {
	return "./templates/views/accounts/confirm_magic"
}

func (c *ConfirmMagicView) Layout() string {
	return "./templates/layouts/index.html"
}

func (c *ConfirmMagicView) OnRequest(w http.ResponseWriter, r *http.Request) (pwc.Status, pwc.Data) {
	if r.Method != "GET" {
		return pwc.Status{Code: 405}, nil
	}
	token := chi.URLParam(r, "token")
	err := c.Auth.LoginWithPasswordlessToken(w, r, token)
	if err != nil {
		return pwc.Status{Code: 200}, nil
	}
	redirectTo := "/app"
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	return pwc.Status{Code: 200}, nil
}
