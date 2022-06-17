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

func (f *ForgotView) OnPatch(event fir.Event) (fir.Patchset, error) {
	switch event.ID {
	case "account/forgot":
		req := new(ProfileRequest)
		if err := event.DecodeParams(req); err != nil {
			return nil, err
		}

		if err := f.Auth.Recovery(event.RequestContext(), req.Email); err != nil {
			return nil, err
		}

		return fir.Patchset{fir.Store{
			Name: "forgot",
			Data: map[string]any{"recovery_sent": true},
		}}, nil
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	return nil, nil
}
