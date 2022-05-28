package accounts

import (
	"log"

	"github.com/adnaan/authn"
	pwc "github.com/adnaan/pineview/controller"
)

type ForgotView struct {
	pwc.DefaultView
	Auth *authn.API
}

func (f *ForgotView) Content() string {
	return "./templates/views/accounts/forgot"
}

func (f *ForgotView) Layout() string {
	return "./templates/layouts/index.html"
}

func (f *ForgotView) OnLiveEvent(ctx pwc.Context) error {
	switch ctx.Event().ID {
	case "account/forgot":
		return f.SendRecovery(ctx)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", ctx.Event())
	}
	return nil
}

func (f *ForgotView) SendRecovery(ctx pwc.Context) error {
	ctx.Store().UpdateProp("show_loading_modal", true)
	defer func() {
		ctx.Store().UpdateProp("show_loading_modal", false)
	}()
	req := new(ProfileRequest)
	if err := ctx.Event().DecodeParams(req); err != nil {
		return err
	}

	if err := f.Auth.Recovery(ctx.Request().Context(), req.Email); err != nil {
		return err
	}

	ctx.Store("forgot").UpdateProp("recovery_sent", true)
	return nil
}
