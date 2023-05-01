package dom

import (
	"html/template"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/livefir/fir/internal/eventstate"
	"golang.org/x/exp/slices"
	"k8s.io/klog/v2"
)

type Event struct {
	Type   *string `json:"type,omitempty"`
	Target *string `json:"target,omitempty"`
	Detail any     `json:"detail,omitempty"`
	// Private fields
	ID    string          `json:"-"`
	State eventstate.Type `json:"-"`
}

func RouteBindings(id string, tmpl *template.Template) *Bindings {
	return &Bindings{
		id:                id,
		tmpl:              tmpl,
		eventTemplates:    make(map[string]map[string]struct{}),
		RWMutex:           &sync.RWMutex{},
		templateNameRegex: regexp.MustCompile(`^[ A-Za-z0-9\-:]*$`),
	}
}

type Bindings struct {
	id                string
	tmpl              *template.Template
	eventTemplates    map[string]map[string]struct{}
	templateNameRegex *regexp.Regexp
	*sync.RWMutex
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
					klog.Errorf(`
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
						klog.Errorf(`
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
						klog.Errorf(`
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
					templates = make(map[string]struct{})
				}

				if !b.templateNameRegex.MatchString(templateName) {
					klog.Errorf("error: invalid template name in event binding: only hyphen(-) and colon(:) are allowed: %v\n", templateName)
					continue
				}

				templates[templateName] = struct{}{}

				//fmt.Printf("eventID: %s, blocks: %v\n", eventID, blocks)
				b.eventTemplates[eventID] = templates
			}
		}

	})

}

func (b *Bindings) TemplateNames(eventIDWithState string) []string {
	b.RLock()
	defer b.RUnlock()
	var templateNames []string
	for k := range b.eventTemplates[eventIDWithState] {
		templateNames = append(templateNames, k)
	}
	return templateNames
}
