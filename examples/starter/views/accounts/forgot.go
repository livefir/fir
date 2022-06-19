package accounts

import (
	"log"

	"github.com/adnaan/authn"
	"github.com/adnaan/fir"
)

type ForgotView struct {
	fir.DefaultView
	Auth *authn.API
}

func (f *ForgotView) Content() string {
	return "./templates/views/accounts/forgot"
}

func (f *ForgotView) Layout() string {
	return "./templates/layouts/index.html"
}

func (f *ForgotView) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "account/forgot":
		req := new(ProfileRequest)
		if err := event.DecodeParams(req); err != nil {
			return nil
		}

		if err := f.Auth.Recovery(event.RequestContext(), req.Email); err != nil {
			return nil
		}

		return fir.Patchset{fir.Store{
			Name: "forgot",
			Data: map[string]any{"recovery_sent": true},
		}}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	return nil
}
