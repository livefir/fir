package account

import (
	"github.com/adnaan/autobahn/models"
	"github.com/adnaan/fir"
)

type View struct {
	DB *models.Client
	fir.DefaultView
}

func (v *View) Content() string {
	return "./views/account"
}

func (v *View) Layout() string {
	return "./templates/layouts/app.html"
}
