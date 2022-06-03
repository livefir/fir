package accounts

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/adnaan/authn"
	fir "github.com/adnaan/fir/controller"
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

func (c *ConfirmEmailChangeView) OnRequest(w http.ResponseWriter, r *http.Request) (fir.Status, fir.Data) {
	if r.Method != "GET" {
		return fir.Status{Code: 405}, nil
	}
	token := chi.URLParam(r, "token")
	userID, _ := r.Context().Value(authn.AccountIDKey).(string)
	acc, err := c.Auth.GetAccount(r.Context(), userID)
	if err != nil {
		log.Printf("confirm change email: GetAccount err %v", err)
		return fir.Status{Code: 200}, nil
	}

	if err := acc.ConfirmEmailChange(r.Context(), token); err != nil {
		log.Printf("confirm change email: ConfirmEmailChange err %v", err)
		return fir.Status{Code: 200}, nil
	}

	redirectTo := "/account"
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	return fir.Status{Code: 200}, nil
}
