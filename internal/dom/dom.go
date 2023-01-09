package dom

import (
	"html/template"
	"io"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/golang/glog"
	"github.com/livefir/fir/pubsub"
	"golang.org/x/exp/slices"
)

type EventState string

const (
	OK      EventState = "ok"
	Error   EventState = "error"
	Pending EventState = "pending"
	Done    EventState = "done"
)

type Event struct {
	Type   *string `json:"type,omitempty"`
	Target *string `json:"target,omitempty"`
	Detail any     `json:"detail,omitempty"`
	// Private fields
	ID    string     `json:"-"`
	State EventState `json:"-"`
}

func NewBindings(id string) Bindings {
	return Bindings{
		id:             id,
		eventTemplates: make(map[string][]string),
	}
}

type Bindings struct {
	id             string
	eventTemplates map[string][]string
	sync.RWMutex
}

func (b *Bindings) Add(rd io.Reader) {
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

func (b *Bindings) Events(pubsubEvent pubsub.Event, tmpl *template.Template) []Event {

	return nil
}
