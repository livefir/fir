package fir

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang/glog"
)

// func replaceFirErrors(ctx Context) (func(err error) []patch.Patch, func() []patch.Patch) {
// 	eventIdName := fmt.Sprintf("%s%s", firErrorPrefix, ctx.event.ID)
// 	eventNameSelector := fmt.Sprintf("#%s", eventIdName)
// 	routeName := "fir-err-route"
// 	routeNameSelector := fmt.Sprintf("#%s", routeName)
// 	return func(err error) []patch.Patch {
// 			errs := map[string]any{ctx.event.ID: err.Error(), "route": err.Error()}
// 			return []patch.Patch{
// 				patch.Replace(eventNameSelector, patch.Block(eventIdName, map[string]any{"fir": newRouteContext(ctx, errs)})),
// 				patch.Replace(routeNameSelector, patch.Block(routeName, map[string]any{"fir": newRouteContext(ctx, errs)}))}
// 		}, func() []patch.Patch {
// 			errs := map[string]any{ctx.event.ID: nil, "route": nil}
// 			return []patch.Patch{
// 				patch.Replace(eventNameSelector, patch.Block(eventIdName, map[string]any{"fir": newRouteContext(ctx, errs)})),
// 				patch.Replace(routeNameSelector, patch.Block(routeName, map[string]any{"fir": newRouteContext(ctx, errs)}))}
// 		}
// }

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

func userError(ctx Context, err error) error {
	userError := err
	glog.Errorf("ctx %+v , error: %v\n", ctx.event.ID, err)
	if wrappedUserError := errors.Unwrap(err); wrappedUserError != nil {
		userError = wrappedUserError
	}
	return userError
}
