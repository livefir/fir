package accounts

import (
	"log"

	"github.com/adnaan/authn"
	pwc "github.com/adnaan/fir/controller"
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

func (f *ForgotView) OnEvent(s pwc.Socket) error {
	switch s.Event().ID {
	case "account/forgot":
		return f.SendRecovery(s)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}

func (f *ForgotView) SendRecovery(s pwc.Socket) error {
	s.Store().UpdateProp("show_loading_modal", true)
	defer func() {
		s.Store().UpdateProp("show_loading_modal", false)
	}()
	req := new(ProfileRequest)
	if err := s.Event().DecodeParams(req); err != nil {
		return err
	}

	if err := f.Auth.Recovery(s.Request().Context(), req.Email); err != nil {
		return err
	}

	s.Store("forgot").UpdateProp("recovery_sent", true)
	return nil
}
