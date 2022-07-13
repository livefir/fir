package accounts

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/adnaan/authn"
	"github.com/adnaan/fir"
)

type ConfirmEmailChangeView struct {
	fir.DefaultView
	Auth *authn.API
}

func (c *ConfirmEmailChangeView) Content() string {
	return "./templates/views/accounts/confirm_email_change"
}

func (c *ConfirmEmailChangeView) Layout() string {
	return "./templates/layouts/app.html"
}

func (c *ConfirmEmailChangeView) OnGet(w http.ResponseWriter, r *http.Request) fir.Page {
	token := chi.URLParam(r, "token")
	userID, _ := r.Context().Value(authn.AccountIDKey).(string)
	acc, err := c.Auth.GetAccount(r.Context(), userID)
	if err != nil {
		return fir.PageError(err, "failed to get account")
	}

	if err := acc.ConfirmEmailChange(r.Context(), token); err != nil {
		return fir.PageError(err, "failed to confirm email change")
	}

	redirectTo := "/account"
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	return fir.Page{}
}
