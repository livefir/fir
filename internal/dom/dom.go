package dom

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/golang/glog"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/pubsub"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"golang.org/x/exp/slices"
)

type Event struct {
	Type   *string `json:"type,omitempty"`
	Target *string `json:"target,omitempty"`
	Detail any     `json:"detail,omitempty"`
	// Private fields
	ID    string          `json:"-"`
	State eventstate.Type `json:"-"`
}

func RouteBindings(id string, tmpl *template.Template) Bindings {
	return Bindings{
		id:             id,
		tmpl:           tmpl,
		eventTemplates: make(map[string][]string),
	}
}

type Bindings struct {
	id             string
	tmpl           *template.Template
	eventTemplates map[string][]string
	sync.RWMutex
}

func (b *Bindings) AddFile(rd io.Reader) {
	b.Lock()
	defer b.Unlock()

	doc, err := goquery.NewDocumentFromReader(rd)
	if err != nil {
		panic(err)
	}
	doc.Find("*").Each(func(_ int, node *goquery.Selection) {
		for _, a := range node.Get(0).Attr {

			if strings.HasPrefix(a.Key, "@fir:") || strings.HasPrefix(a.Key, "x-on:fir:") {

				eventns := strings.TrimPrefix(a.Key, "@fir:")
				eventns = strings.TrimPrefix(eventns, "x-on:fir:")
				eventnsParts := strings.SplitN(eventns, ".", -1)
				if len(eventnsParts) > 3 {
					glog.Errorf(`
					error: invalid event namespace: %s. 
					must be of the format => @fir:<event>:<ok|error>:<block-name|optional>`, eventns)
					continue
				}

				if len(eventnsParts) > 0 {
					eventns = eventnsParts[0]
				}

				eventnsParts = strings.SplitN(eventns, ":", -1)
				if len(eventnsParts) == 0 {
					continue
				}
				eventID := eventnsParts[0]
				if len(eventnsParts) >= 2 {
					if !slices.Contains([]string{"ok", "error", "pending", "done"}, eventnsParts[1]) {
						glog.Errorf(`
						error: invalid event namespace: %s. 
						it must be of the format => 
						@fir:<event>:<ok|error>:<block|optional> or
						@fir:<event>:<pending|done>`, eventns)
						continue
					}
					if len(eventnsParts) == 2 {
						if eventnsParts[1] == "pending" || eventnsParts[1] == "done" {
							continue
						}
					}
					eventID = strings.Join(eventnsParts[0:2], ":")
				}

				templateName := "-"
				if len(eventnsParts) == 3 {
					if !slices.Contains([]string{"ok", "error"}, eventnsParts[1]) {
						glog.Errorf(`
						error: invalid event namespace: %s. 
						it must be of the format => 
						@fir:<event>:<ok|error>:<block|optional> or
						@fir:<event>:<pending|done>.
						<block> cannot be set for <pending|done> since they are client only`, eventns)
						continue
					}
					templateName = eventnsParts[2]
				}

				templates, ok := b.eventTemplates[eventID]
				if !ok {
					templates = []string{}
				}

				templates = append(templates, templateName)

				//fmt.Printf("eventID: %s, blocks: %v\n", eventID, blocks)
				b.eventTemplates[eventID] = templates

			}
		}

	})

}

func (b *Bindings) Events(pubsubEvent pubsub.Event) []Event {
	b.RLock()
	eventIDWithState := fmt.Sprintf("%s:%s", *pubsubEvent.ID, pubsubEvent.State)
	templates, ok := b.eventTemplates[eventIDWithState]
	b.RUnlock()
	if !ok {
		return []Event{}
	}
	var events []Event
	for _, templateName := range templates {
		value, err := buildTemplateValue(b.tmpl, templateName, pubsubEvent.Detail)
		if err != nil {
			glog.Errorf("Bindings.Events buildTemplateValue error: %s", err)
			continue
		}
		eventType := fmt.Sprintf("fir:%s:%s", eventIDWithState, templateName)
		events = append(events, Event{
			ID:     b.id,
			State:  pubsubEvent.State,
			Type:   &eventType,
			Target: pubsubEvent.Target,
			Detail: value,
		})
	}

	return events
}

func buildTemplateValue(t *template.Template, name string, data any) (string, error) {
	var buf bytes.Buffer
	defer buf.Reset()
	if name == "_fir_html" {
		buf.WriteString(data.(string))
	} else {
		t.Option("missingkey=zero")
		err := t.ExecuteTemplate(&buf, name, data)
		if err != nil {
			return "", err
		}
	}

	m := minify.New()
	m.Add("text/html", &html.Minifier{})
	r := m.Reader("text/html", &buf)
	var buf1 bytes.Buffer
	defer buf1.Reset()
	_, err := io.Copy(&buf1, r)
	if err != nil {
		return "", err
	}
	value := buf1.String()
	return value, nil
}
