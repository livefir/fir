package main

import (
	"log"
	"net/http"

	pwc "github.com/adnaan/pineview/controller"
)

type LayoutView struct {
	pwc.DefaultView
}

func (l *LayoutView) Layout() string {
	return "templates/layouts/index.html"
}

type HomeView struct {
	LayoutView
}

func (h *HomeView) Content() string {
	return "templates/views/home"
}

type HelpView struct {
	LayoutView
}

func (h *HelpView) Content() string {
	return "templates/views/help"
}

type SettingsView struct {
	LayoutView
}

func (h *SettingsView) Content() string {
	return "templates/views/settings"
}

func main() {
	glvc := pwc.Websocket("pineview-layout", pwc.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(&HomeView{}))
	http.Handle("/help", glvc.Handler(&HelpView{}))
	http.Handle("/settings", glvc.Handler(&SettingsView{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
