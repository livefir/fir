package app

import (
	"log"
	"net/http"

	"github.com/adnaan/authn"
	pwc "github.com/adnaan/pineview/controller"
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

func (d *DashboardView) OnLiveEvent(ctx pwc.Context) error {
	switch ctx.Event().ID {
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", ctx.Event())
	}
	return nil
}

func (d *DashboardView) OnMount(w http.ResponseWriter, r *http.Request) (pwc.Status, pwc.M) {
	return pwc.Status{Code: 200}, pwc.M{
		"is_logged_in": true,
	}
}
