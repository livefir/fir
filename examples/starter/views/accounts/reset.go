package accounts

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/adnaan/authn"
	pwc "github.com/adnaan/fir/controller"
)

type ResetView struct {
	pwc.DefaultView
	Auth *authn.API
}

func (rv *ResetView) Content() string {
	return "./templates/views/accounts/reset"
}

func (rv *ResetView) Layout() string {
	return "./templates/layouts/index.html"
}

func (rv *ResetView) OnLiveEvent(ctx pwc.Context) error {
	switch ctx.Event().ID {
	case "account/reset":
		return rv.Reset(ctx)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", ctx.Event())
	}
	return nil
}

func (rv *ResetView) OnMount(w http.ResponseWriter, r *http.Request) (pwc.Status, pwc.M) {
	token := chi.URLParam(r, "token")
	return pwc.Status{Code: 200}, pwc.M{
		"token": token,
	}
}

type ResetReq struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Token           string `json:"token"`
}

func (rv *ResetView) Reset(ctx pwc.Context) error {
	ctx.Store().UpdateProp("show_loading_modal", true)
	defer func() {
		ctx.Store().UpdateProp("show_loading_modal", false)
	}()
	r := new(ResetReq)
	if err := ctx.Event().DecodeParams(r); err != nil {
		return err
	}
	if r.ConfirmPassword != r.Password {
		return fmt.Errorf("%w", errors.New("passwords don't match"))
	}
	if err := rv.Auth.ConfirmRecovery(ctx.Request().Context(), r.Token, r.Password); err != nil {
		return err
	}
	ctx.Store("reset").UpdateProp("password_reset", true)
	return nil
}
