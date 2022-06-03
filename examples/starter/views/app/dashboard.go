package app

import (
	"log"
	"net/http"

	"github.com/adnaan/authn"
	fir "github.com/adnaan/fir/controller"
)

type DashboardView struct {
	fir.DefaultView
	Auth *authn.API
}

func (d *DashboardView) Content() string {
	return "./templates/views/app"
}

func (d *DashboardView) Layout() string {
	return "./templates/layouts/app.html"
}

func (d *DashboardView) OnEvent(s fir.Socket) error {
	switch s.Event().ID {
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}

func (d *DashboardView) OnRequest(w http.ResponseWriter, r *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"is_logged_in": true,
	}
}
