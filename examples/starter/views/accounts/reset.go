package accounts

import (
	"errors"
	"fmt"
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

func (rv *ResetView) OnEvent(s fir.Socket) error {
	switch s.Event().ID {
	case "account/reset":
		return rv.Reset(s)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
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

func (rv *ResetView) Reset(s fir.Socket) error {
	s.Store().UpdateProp("show_loading_modal", true)
	defer func() {
		s.Store().UpdateProp("show_loading_modal", false)
	}()
	r := new(ResetReq)
	if err := s.Event().DecodeParams(r); err != nil {
		return err
	}
	if r.ConfirmPassword != r.Password {
		return fmt.Errorf("%w", errors.New("passwords don't match"))
	}
	if err := rv.Auth.ConfirmRecovery(s.Request().Context(), r.Token, r.Password); err != nil {
		return err
	}
	s.Store("reset").UpdateProp("password_reset", true)
	return nil
}
