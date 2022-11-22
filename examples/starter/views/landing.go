package views

import (
	"net/http"

	"github.com/adnaan/authn"
	"github.com/adnaan/fir"
)

type LandingView struct {
	fir.DefaultView
	Auth *authn.API
}

func (l *LandingView) Content() string {
	return "./templates/views/landing"
}

func (l *LandingView) Layout() string {
	return "./templates/layouts/index.html"
}

func (l *LandingView) OnPost(_ http.ResponseWriter, r *http.Request) fir.Page {
	if _, err := l.Auth.CurrentAccount(r); err != nil {
		return fir.Page{}
	}
	return fir.Page{
		Data: map[string]any{
			"is_logged_in": true,
		}}
}
