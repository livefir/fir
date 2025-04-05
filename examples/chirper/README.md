# Chirper: a simple real-time twitter clone

The example demonstrates: progressive enhancement, form validation and real-time changes.

## Start without javascript
[index_no_js.html](./index_no_js.html) is a plain html file which is handled by the route function [NoJSIndex](index.go#NoJSIndex). 


### List chirps

Whenever the page loads, the bound `OnLoad` function(`loadChirps`) is automatically invoked. The returned data(`ctx.Data`) on a successful loading of chirps from the database is used to re-render the entire page([index_no_js.html](./index_no_js.html)
). 

https://github.com/livefir/fir/blob/919cb9ae8dc0cac66ba2d07b15d09cad4e92f76a/examples/chirper/index.go#L100

https://github.com/livefir/fir/blob/919cb9ae8dc0cac66ba2d07b15d09cad4e92f76a/examples/chirper/index.go#L21-L32

An error returned from the `loadChirps` can be rendered on the page using `fir.Errror` template function : `{{ fir.Error "onload" }}`.

https://github.com/livefir/fir/blob/919cb9ae8dc0cac66ba2d07b15d09cad4e92f76a/examples/chirper/index_no_js.html#L30

Listing the chirps itself is just standard html/template.

https://github.com/livefir/fir/blob/919cb9ae8dc0cac66ba2d07b15d09cad4e92f76a/examples/chirper/index_no_js.html#L33-L60


### Create chirp

We will create a `chirp`by submitting a html form with an action of format `action="?event=event-name"` to invoke the bound `onEvent`function on the server. The form method must use `POST` to invoke `OnEvent`otherwise `OnLoad` will be called with form data passed as query params. Please note that there is no forward slash(/) in the form action to ensure that the form is submitted to the current path.


https://github.com/livefir/fir/blob/919cb9ae8dc0cac66ba2d07b15d09cad4e92f76a/examples/chirper/index_no_js.html#L18-L27


The event `create-chirp` is bound to an event handler of type [func(ctx RouteContext) error](https://pkg.go.dev/github.com/livefir/fir#OnEventFunc).

https://github.com/livefir/fir/blob/5040291ae4c2de65e379bb904ca64d7614b6e707/examples/chirper/index.go#L101

https://github.com/livefir/fir/blob/919cb9ae8dc0cac66ba2d07b15d09cad4e92f76a/examples/chirper/index.go#L34-L52


Within the `createChirp` function, [RouteContext.Bind](https://pkg.go.dev/github.com/livefir/fir#RouteContext.Bind) is used to bind the form data to the request struct and return errors to render failures on the html page. In the html page, the returned error for an event can be rendered using the `fir.Errror` template function : `{{ fir.Error "create-chirp" }}`. `createChirp` also returns a [FieldError](https://pkg.go.dev/github.com/livefir/fir#RouteContext.FieldError) to indicate the specific field which failed validation. The field error can be referenced like so: `{{ fir.Error "create-chirp.body" }}` where `create-chirp`is the event id and `body`is the form field. 

Since the current page involves no javascript, the page reloads which re-renders the whole page using the data from `loadChirps`as described in the section above.

### Like and Delete chirp

To increment the like count `like-count` and to delete the chirp itself `delete-chirp` events are submitted using an html form.

https://github.com/livefir/fir/blob/5040291ae4c2de65e379bb904ca64d7614b6e707/examples/chirper/index_no_js.html#L36-L57

https://github.com/livefir/fir/blob/5040291ae4c2de65e379bb904ca64d7614b6e707/examples/chirper/index.go#L102-L103

https://github.com/livefir/fir/blob/5040291ae4c2de65e379bb904ca64d7614b6e707/examples/chirper/index.go#L54-L92

Just like during `createChirp`, errors can be rendered using `fir.Error` while the page is re-rendered with data returned from `loadChirps`


See it in action:

```
go run .
open http://localhost:9867/nojs
```

The above page works even if javascript is disabled in the browser.

## Enhance with the alpinejs client

[index.html](./index.html) is an html file with javascript sprinkled using the alpinejs client. It is handled by the route function [Index](index.go#Index).

We want to enhance the html page to avoid reloads and re-render only the changed parts of the DOM.


### Enhance create chirp

To stop the page from reloading, we will prevent the form submission and submit the event over the wire. Depending on whether websocket is enabled on the server, the alpinejs plugin will send the event as a websocket message or as a POST request. We use alpinejs [x-on](https://alpinejs.dev/directives/on) directive to bind the form's submit event to Fir's [custom magic function](https://alpinejs.dev/advanced/extending#magic-functions) `$fir.submit`

https://github.com/livefir/fir/blob/e38ea115dfcecbb4b890d94acd49e9565bdf2146/examples/chirper/index.html#L28


See it in action:

```
open http://localhost:9867
```

Open two tabs at: http://localhost:9867. Add a chirp in one and see it broadcasted to the second tab instantly.