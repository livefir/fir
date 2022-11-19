package fir

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func onPatchEvent(w http.ResponseWriter, r *http.Request, v *viewHandler) {
	v.reloadTemplates()
	var event Event
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if decoder.More() {
		http.Error(w, "unknown fields in request body", http.StatusBadRequest)
		return
	}
	if event.ID == "" {
		http.Error(w, "event id is missing", http.StatusBadRequest)
		return
	}
	event.requestContext = r.Context()
	patchset := getEventPatchset(event, v.view)
	channel := *v.cntrl.channelFunc(r, v.view.ID())
	err = v.cntrl.pubsub.Publish(r.Context(), channel, patchset)
	if err != nil {
		log.Printf("[onPatchEvent] error publishing patch: %v\n", err)
	}
	w.Write(buildPatchOperations(v.viewTemplate, patchset))
}

func onRequest(w http.ResponseWriter, r *http.Request, v *viewHandler) {
	v.reloadTemplates()

	var err error
	var page Page

	if r.Method == "POST" {
		page = v.view.OnPost(w, r)
	} else {
		page = v.view.OnGet(w, r)
	}
	v.mountData = page.Data
	if v.mountData == nil {
		v.mountData = make(Data)
	}

	v.mountData["app_name"] = v.cntrl.name
	v.mountData["fir"] = &PageContext{
		Name:    v.cntrl.name,
		URLPath: r.URL.Path,
	}

	page.Data = v.mountData

	if page.Code == 0 {
		page.Code = http.StatusOK
	}
	if page.Message == "" {
		page.Message = http.StatusText(page.Code)
	}

	if page.Code > 299 {
		log.Printf("page error: %v\n", page.Error)
		onRequestError(w, r, v, &page)
		return
	}

	v.viewTemplate.Option("missingkey=zero")
	var buf bytes.Buffer
	err = v.viewTemplate.Execute(&buf, page.Data)
	if err != nil {
		log.Printf("OnGet viewTemplate.Execute error:  %v", err)
		onRequestError(w, r, v, nil)
	}
	if v.cntrl.debugLog {
		// log.Printf("OnGet render view %+v, with data => \n %+v\n",
		// 	v.view.Content(), getJSON(page.Data))
	}

	w.WriteHeader(page.Code)
	w.Write(buf.Bytes())

}

func onRequestError(w http.ResponseWriter, r *http.Request, v *viewHandler, page *Page) {
	errorPage := v.errorView.OnGet(w, r)
	if page == nil {
		page = &errorPage
	}
	v.mountData = page.Data
	if v.mountData == nil {
		v.mountData = make(Data)
	}
	v.mountData["statusCode"] = page.Code
	v.mountData["statusMessage"] = page.Message

	page.Data = v.mountData

	v.viewTemplate.Option("missingkey=zero")
	var buf bytes.Buffer
	err := v.errorViewTemplate.Execute(&buf, page.Data)
	if err != nil {
		log.Printf("err rendering error template: %v\n", err)
		_, errWrite := w.Write([]byte("Something went wrong"))
		if errWrite != nil {
			panic(errWrite)
		}
	}

	if v.cntrl.debugLog {
		log.Printf("OnGet render error view %+v, with data => \n %+v\n",
			v.view.Content(), getJSON(page.Data))
	}

	w.WriteHeader(page.Code)
	w.Write(buf.Bytes())

}
