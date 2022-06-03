package views

import (
	fir "github.com/adnaan/fir/controller"
)

type NotfoundView struct {
	fir.DefaultView
}

func (n *NotfoundView) Content() string {
	return "./templates/404.html"
}

func (n *NotfoundView) Layout() string {
	return "./templates/layouts/error.html"
}
