package fir

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
)

// Op is the type of patch operation
type Op string

const (
	morph       Op = "morph"
	after       Op = "after"
	before      Op = "before"
	appendOp    Op = "append"
	prepend     Op = "prepend"
	remove      Op = "remove"
	reload      Op = "reload"
	updateStore Op = "store"
	resetForm   Op = "resetForm"
	navigate    Op = "navigate"
)

// Patch is an interface for all patch operations
type Patch struct {
	// Op is the type of patch operation
	Op Op `json:"op"`
	// Selector is the css selector for the element to patch
	Selector *string `json:"selector,omitempty"`
	// Value is the value for the patch operation
	Value any `json:"value,omitempty"`
}

func Morph(selector, templateName string, data any) Patch {
	return Patch{
		Op:       morph,
		Selector: &selector,
		Value:    map[string]any{"name": templateName, "data": data},
	}
}

func After(selector, templateName string, data any) Patch {
	return Patch{
		Op:       after,
		Selector: &selector,
		Value:    map[string]any{"name": templateName, "data": data},
	}
}

func Before(selector, templateName string, data any) Patch {
	return Patch{
		Op:       before,
		Selector: &selector,
		Value:    map[string]any{"name": templateName, "data": data},
	}
}

func Append(selector, templateName string, data any) Patch {
	return Patch{
		Op:       appendOp,
		Selector: &selector,
		Value:    map[string]any{"name": templateName, "data": data},
	}
}

func Prepend(selector, templateName string, data any) Patch {
	return Patch{
		Op:       prepend,
		Selector: &selector,
		Value:    map[string]any{"name": templateName, "data": data},
	}
}

func Remove(selector string) Patch {
	return Patch{
		Op:       morph,
		Selector: &selector,
	}
}

func Reload() Patch {
	return Patch{
		Op: morph,
	}
}

func Store(name string, data any) Patch {
	return Patch{
		Op:    updateStore,
		Value: data,
	}
}

func ResetForm(selector string) Patch {
	return Patch{
		Op:       resetForm,
		Selector: &selector,
	}
}

func Navigate(url string) Patch {
	return Patch{
		Op:    navigate,
		Value: url,
	}
}

func buildPatchOperations(t *template.Template, patchset []Patch) []byte {
	var renderedPatchset []Patch
	firErrorPatchExists := false
	for _, p := range patchset {
		switch p.Op {
		case updateStore, navigate, resetForm, reload, remove:
			renderedPatchset = append(renderedPatchset, p)
		case morph, after, before, appendOp, prepend:
			tmpl, ok := p.Value.(map[string]any)
			if !ok {
				log.Printf("[buildPatchOperations] invalid patch template data: %v", p.Value)
				continue
			}

			if *p.Selector == "#fir-error" {
				firErrorPatchExists = true
			}

			var err error
			p.Value, err = buildTemplateValue(t, tmpl["name"].(string), tmpl["data"])
			if err != nil {
				log.Printf("buildPatchOperations error: %v,%+v \n", err, tmpl)
				continue
			}

			renderedPatchset = append(renderedPatchset, p)
		default:
			continue
		}
	}

	if !firErrorPatchExists {
		// unset error patch
		firError := "#fir-error"
		tmplVal, err := buildTemplateValue(t, "fir-error", nil)
		if err == nil {
			renderedPatchset = append([]Patch{{
				Op:       morph,
				Selector: &firError,
				Value:    tmplVal,
			}}, renderedPatchset...)
		}
	}

	data, err := json.Marshal(renderedPatchset)
	if err != nil {
		log.Printf("buildPatchOperations marshal error: %+v, %v \n", renderedPatchset, err)
		return nil
	}
	return data
}

func buildTemplateValue(t *template.Template, name string, data any) (string, error) {
	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return "", err
	}
	value := buf.String()
	buf.Reset()
	return value, nil
}
