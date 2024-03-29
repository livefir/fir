# Chirper: a simple real-time twitter clone

The example demonstrates: progressive enhancement, form validation and real-time changes.

## Start without javascript
[index_no_js.html](./index_no_js.html) is a plain html file which is handled by the route function [NoJSIndex](index.go#NoJSIndex). See it in action:

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