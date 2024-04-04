# Chirper: a simple real-time twitter clone

The example demonstrates: progressive enhancement, form validation and real-time changes.

## Start without javascript
[index_no_js.html](./index_no_js.html) is a plain html file which is handled by the route function [NoJSIndex](index.go#NoJSIndex). 




### Create chirp

We will create a `chirp`by submitting a html form with an action of format `action="?event=event-name"` to invoke the bound `onEvent`function on the server. The form method must use `POST` to invoke `onEvent`otherwise `onLoad` will be called with form data passed as query params. Please note that there is no forward slash(/) in the form action to ensure that the form is submitted to the current path.


https://github.com/livefir/fir/blob/3b25bc50187198b5b0913f1d8611c4aaba36a302/examples/chirper/index_no_js.html#L21-L30


The event `create-chirp` is bound to an event handler of type [func(ctx RouteContext) error](https://pkg.go.dev/github.com/livefir/fir#OnEventFunc).



Within the `createChirp` function, [RouteContext.Bind](https://pkg.go.dev/github.com/livefir/fir#RouteContext.Bind) is used to bind the form data to the request struct and return errors to render failures on the html page. In the html page, the returned error for an event can be rendered using the `fir.Errror` template function : `{{ fir.Error "create-chirp" }}`. `createChirp` also returns a [FieldError](https://pkg.go.dev/github.com/livefir/fir#RouteContext.FieldError) to indicate the specific field which failed validation. The field error can be referenced like so: `{{ fir.Error "create-chirp.body" }}` where `create-chirp`is the event id and `body`is the form field. 









### List chirps

### Like chirp

### Delete chirp


See it in action:

```
go run .
open http://localhost:9867/nojs
```

The above page works even if javascript is disabled in the browser.

## Enhance with the alpinejs client

[index.html](./index.html) is an html file with javascript sprinkled using the alpinejs client. It is handled by the route function [Index](index.go#Index). See it in action:

```
open http://localhost:9867
```

Open two tabs at: http://localhost:9867. Add a chirp in one and see it broadcasted to the second tab instantly.