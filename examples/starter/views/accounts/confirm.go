package accounts

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/adnaan/authn"
	pwc "github.com/adnaan/fir/controller"
)

type ConfirmView struct {
	pwc.DefaultView
	Auth *authn.API
}

func (h *ConfirmView) Content() string {
	return "./templates/views/accounts/confirm"
}

func (h *ConfirmView) Layout() string {
	return "./templates/layouts/index.html"
}

func (h *ConfirmView) OnMount(w http.ResponseWriter, r *http.Request) (pwc.Status, pwc.M) {
	token := chi.URLParam(r, "token")
	err := h.Auth.ConfirmSignupEmail(r.Context(), token)
	if err != nil {
		log.Println("err confirm.onmount", err)
		return pwc.Status{Code: 200}, nil
	}
	return pwc.Status{Code: 200}, pwc.M{
		"confirmed": true,
	}
}
