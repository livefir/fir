package accounts

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/adnaan/authn"
	fir "github.com/adnaan/fir/controller"
)

type ConfirmView struct {
	fir.DefaultView
	Auth *authn.API
}

func (h *ConfirmView) Content() string {
	return "./templates/views/accounts/confirm"
}

func (h *ConfirmView) Layout() string {
	return "./templates/layouts/index.html"
}

func (h *ConfirmView) OnRequest(w http.ResponseWriter, r *http.Request) (fir.Status, fir.Data) {
	token := chi.URLParam(r, "token")
	err := h.Auth.ConfirmSignupEmail(r.Context(), token)
	if err != nil {
		log.Println("err confirm.OnRequest", err)
		return fir.Status{Code: 200}, nil
	}
	return fir.Status{Code: 200}, fir.Data{
		"confirmed": true,
	}
}
