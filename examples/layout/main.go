package main

import (
	"log"
	"net/http"

	fir "github.com/adnaan/fir/controller"
)

type LayoutView struct {
	fir.DefaultView
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
	glvc := fir.Websocket("fir-layout", fir.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(&HomeView{}))
	http.Handle("/help", glvc.Handler(&HelpView{}))
	http.Handle("/settings", glvc.Handler(&SettingsView{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
