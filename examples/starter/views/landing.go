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

func (l *LandingView) OnRequest(_ http.ResponseWriter, r *http.Request) (fir.Status, fir.Data) {
	if r.Method != "GET" {
		return fir.Status{Code: 405}, nil
	}
	if _, err := l.Auth.CurrentAccount(r); err != nil {
		return fir.Status{Code: 200}, nil
	}
	return fir.Status{Code: 200}, fir.Data{
		"is_logged_in": true,
	}
}
