package renderer

import (
"fmt"
"html/template"

"github.com/livefir/fir/internal/dom"
"github.com/livefir/fir/internal/eventstate"
"github.com/patrickmn/go-cache"
"github.com/tdewolff/minify"
"github.com/tdewolff/minify/html"
"github.com/valyala/bytebufferpool"
)

func TargetOrClassName(target *string, className string) *string {
	if target != nil && *target != "" {
		return target
	}
	cls := fmt.Sprintf(".%s", className)
	return &cls
}

func BuildTemplateValue(t *template.Template, templateName string, data any, addAttributesFunc func([]byte) []byte) (string, error) {
	if t == nil {
		return "", nil
	}
	if templateName == "" {
		return "", nil
	}
	dataBuf := bytebufferpool.Get()
	defer bytebufferpool.Put(dataBuf)
	if templateName == "_fir_html" {
		dataBuf.WriteString(data.(string))
	} else {
		err := t.ExecuteTemplate(dataBuf, templateName, data)
		if err != nil {
			return "", err
		}
	}

	m := minify.New()
	m.Add("text/html", &html.Minifier{
		KeepDefaultAttrVals: true,
	})
	rd, err := m.Bytes("text/html", addAttributesFunc(dataBuf.Bytes()))
	if err != nil {
		panic(err)
	}

	return string(rd), nil
}

func IsEmptyEvent(event dom.Event) bool {
	return event.Type == nil && event.Target == nil && event.Key == nil
}

func Fir(parts ...string) *string {
	if len(parts) == 0 {
		return nil
	}
	if len(parts) == 1 {
		return &parts[0]
	}
	result := fmt.Sprintf("%s:%s", parts[0], parts[1])
	return &result
}

func GetUnsetErrorEvents(cch *cache.Cache, sessionID *string, events []dom.Event) []dom.Event {
	if sessionID == nil || cch == nil {
		return nil
	}

	prevErrors := make(map[string]string)

	v, ok := cch.Get(*sessionID)
	if ok {
		prevErrors, ok = v.(map[string]string)
		if !ok {
			panic("fir: cache value is not a map[string]string")
		}
	}

	currErrors := make(map[string]string)
	for _, event := range events {
		if event.Type == nil {
			continue
		}
		if event.State != eventstate.Error {
			continue
		}
		currErrors[*event.Type] = *event.Target
	}

	cch.Set(*sessionID, currErrors, cache.DefaultExpiration)

	var newErrorEvents []dom.Event
	for k, v := range prevErrors {
		k := k
		v := v
		eventType := &k
		target := v
		if _, ok := currErrors[*eventType]; ok {
			continue
		}
		newErrorEvents = append(newErrorEvents, dom.Event{
Type:   eventType,
Target: &target,
		})
	}

	return newErrorEvents
}
