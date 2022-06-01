package app

import (
	"log"
	"net/http"

	"github.com/adnaan/authn"
	pwc "github.com/adnaan/fir/controller"
)

type DashboardView struct {
	pwc.DefaultView
	Auth *authn.API
}

func (d *DashboardView) Content() string {
	return "./templates/views/app"
}

func (d *DashboardView) Layout() string {
	return "./templates/layouts/app.html"
}

func (d *DashboardView) OnEvent(s pwc.Socket) error {
	switch s.Event().ID {
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}

func (d *DashboardView) OnRequest(w http.ResponseWriter, r *http.Request) (pwc.Status, pwc.Data) {
	return pwc.Status{Code: 200}, pwc.Data{
		"is_logged_in": true,
	}
}
