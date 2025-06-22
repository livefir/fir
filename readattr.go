package fir

import (
	"bytes"

	"github.com/livefir/fir/internal/firattr"
	"github.com/sourcegraph/conc/pool"
	"golang.org/x/net/html"
)

func readAttributes(fi fileInfo) fileInfo {
	doc, err := html.Parse(bytes.NewReader(fi.content))
	if err != nil {
		panic(err)
	}

	attributes := firattr.FirAttributes(doc)
	resultPool := pool.NewWithResults[eventTemplates]()
	for _, attr := range attributes {
		attr := attr
		resultPool.Go(func() eventTemplates {
			return eventTemplatesFromAttr(attr)
		})
	}
	evtArr := resultPool.Wait()
	evt := make(eventTemplates)
	for _, evtMap := range evtArr {
		evt = deepMergeEventTemplates(evt, evtMap)
	}

	return fileInfo{
		name:           fi.name,
		content:        fi.content,
		err:            fi.err,
		blocks:         fi.blocks,
		eventTemplates: evt,
	}
}

func eventTemplatesFromAttr(attr html.Attribute) eventTemplates {
	firattrEvt := firattr.EventTemplatesFromAttr(attr, templateNameRegex)

	// Convert firattr.EventTemplates to main package eventTemplates
	evt := make(eventTemplates)
	for eventID, templates := range firattrEvt {
		mainTemplates := make(eventTemplate)
		for templateName := range templates {
			mainTemplates[templateName] = struct{}{}
		}
		evt[eventID] = mainTemplates
	}

	return evt
}
