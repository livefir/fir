# Chirper: a simple real-time twitter clone

The example demonstrates: progressive enhancement, form validation and real-time changes.

## Start without javascript
[index_no_js.html](./index_no_js.html) is a plain html file which is handled by the route function [NoJSIndex](index.go#NoJSIndex). 




### Create chirp

We will create a `chirp`by submitting a html form with an action of format `action="/?event=event-name"` to invoke the bound `onEvent`function on the server. The form method must use `POST` to invoke `onEvent`otherwise `onLoad` will be called with form data passed as query params.


https://github.com/livefir/fir/blob/main/examples/chirper/index_no_js.html#L21-L30



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