package accounts

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/adnaan/authn"
	"github.com/adnaan/fir"
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

func (c *ConfirmMagicView) OnGet(w http.ResponseWriter, r *http.Request) (fir.Status, fir.Data) {
	token := chi.URLParam(r, "token")
	err := c.Auth.LoginWithPasswordlessToken(w, r, token)
	if err != nil {
		return fir.Status{Code: 200}, nil
	}
	redirectTo := "/app"
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	return fir.Status{Code: 200}, nil
}
