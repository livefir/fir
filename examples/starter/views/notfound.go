package views

import (
	pwc "github.com/adnaan/pineview/controller"
)

type NotfoundView struct {
	pwc.DefaultView
}

func (n *NotfoundView) Content() string {
	return "./templates/404.html"
}

func (n *NotfoundView) Layout() string {
	return "./templates/layouts/error.html"
}
