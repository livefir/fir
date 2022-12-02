package fir

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang/glog"
)

func MorphError(name string) (func(err error) Patch, func() Patch) {
	selector := fmt.Sprintf("#%s", name)
	return func(err error) Patch {
			return Morph(selector, Block(name, M{name: err}))
		}, func() Patch {
			return Morph(selector, Block(name, M{name: ""}))
		}
}

func morphFirErrors(ctx Context) (func(err error) []Patch, func() []Patch) {
	eventIdName := fmt.Sprintf("fir-error-%s", ctx.event.ID)
	eventNameSelector := fmt.Sprintf("#%s", eventIdName)
	routeName := "fir-err-route"
	routeNameSelector := fmt.Sprintf("#%s", routeName)
	return func(err error) []Patch {
			errs := map[string]any{ctx.event.ID: err.Error(), "route": err.Error()}
			return []Patch{
				Morph(eventNameSelector, Block(eventIdName, M{"fir": newRouteContext(ctx, errs)})),
				Morph(routeNameSelector, Block(routeName, M{"fir": newRouteContext(ctx, errs)}))}
		}, func() []Patch {
			errs := map[string]any{ctx.event.ID: nil, "route": nil}
			return []Patch{
				Morph(eventNameSelector, Block(eventIdName, M{"fir": newRouteContext(ctx, errs)})),
				Morph(routeNameSelector, Block(routeName, M{"fir": newRouteContext(ctx, errs)}))}
		}
}

type fieldErrors map[string]error

func (f fieldErrors) toMap() map[string]string {
	m := map[string]string{}
	for field, err := range f {
		m[field] = err.Error()
	}
	return m
}

func (f fieldErrors) Error() string {
	var errs []string
	for field, err := range f {
		errs = append(errs, fmt.Sprintf("%s: %s", field, err.Error()))
	}
	return strings.Join(errs, ", ")
}

func UserError(ctx Context, err error) error {
	userError := err
	glog.Errorf("ctx %+v , error: %v\n", ctx.event.ID, err)
	if wrappedUserError := errors.Unwrap(err); wrappedUserError != nil {
		userError = wrappedUserError
	}
	return userError
}
