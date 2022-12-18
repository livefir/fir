# Introduction

The `@livefir/fir` package is a companion [alpine.js](https://alpinejs.dev/advanced/extending#via-script-tag) plugin for the Go library [pkg.go.dev/github.com/livefir/fir](https://pkg.go.dev/github.com/livefir/fir). Fir is a Go toolkit to build reactive web interfaces using: [Go](https://go.dev/), [html/template](https://pkg.go.dev/html/template) and [alpinejs](https://alpinejs.dev/). 

## Installation

The plugin can be included as a script tag before the `alpine.js` include.

```html
    <head>
        <script
            defer
            src="https://unpkg.com/@livefir/fir@latest/dist/fir.min.js"></script>
        <script
            defer
            src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
    </head>
```


 See [example](https://github.com/livefir/fir#example).