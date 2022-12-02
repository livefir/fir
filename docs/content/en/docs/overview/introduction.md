---
title: "Introduction"
description: ""
lead: ""
date: 2022-11-20T18:05:16+01:00
lastmod: 2022-11-20T18:05:16+01:00
draft: false
images: []
menu:
  docs:
    parent: ""
    identifier: "introduction-c10be8ad434e49b824ec1ba93352f840"
weight: 999
toc: true
---

Fir is a toolkit to build server-rendered HTML apps and progressively enhance them to enable real-time user experiences. The toolkit is meant for developers who want to build real-time web apps using Go, server-rendered HTML(html/template), CSS and sprinkles of declarative javascript(alpine.js). It can be used to build: static websites like a landing page or a blog,  interactive CRUD apps like ticket helpdesks, real-time apps like metrics dashboards or social media streams.

The toolkit consists of a Go server library, an alpine.js plugin, and a CLI. The server library can render HTML over both HTTP and real-time connections(websockets, HTTP/2) while the alpine.js plugin provides a declarative API to enhance plain HTML to support real-time user interaction. The CLI can generate HTML views and SQL ORMs from entgo schemas(similar to Django, Rails, Phoenix frameworks).

The big idea behind Fir is to enhance Go’s standard html template engine with alpine.js to  patch only parts of an HTML on user interaction and without page reloads. A Fir app can start as a simple HTML app which works without javascript depending on browser’s capability(form submissions) to handle user interactions but requires page reloads. This server-rendered HTML app can later be enhanced using sprinkles of declarative javascript provided by the companion alpine.js plugin. After the enhancement, user interactions(clicks, form submits) are sent over a real-time connection(websocket, HTTP/2) and change operations(patching the DOM, updating state, navigation etc.) are applied on the client without page reloads. The server library has the capability to publish these page change operations to all connected clients(multiple devices, multiple tabs etc.) which enables a real-time user experience. When real-time connections are not available, the client falls back to using standard HTTP while still providing user interactions without page reloads.

Consider the following html/template page:

```go
<form name="newTodo" method="post">
 <input name="text" class="input" type="text" required/>
  <button type="submit" value="Submit">Submit</button>
</form>
<div id="todos">
{{ range .todos }}
 <div id="{{.ID}}"> {{.Text}} </div>
{{ end }}
</div>
```

On form submission, server’s template engine hydrates the page template with updated `todos` and renders the new html page as http response.

To avoid a page reload and update only the changed part of the page (i.e `{{range .todos}} .Text {{end}}` ), we need a client javascript function which can fetch the updated part from the server and patch the relevant section of the DOM. Also, the server’s template engine should be capable of re-rendering only the changed part of the template. A Go html template page can be composed of reusable parts enclosed in a `template`or `block` expression. Fir builds on top of Go’s standard capability by providing a way to re-render a `template/block` part and patching the targeted part of the page without a page reload while using only a bit of javascript.

```go
<form name="newTodo" method="post">
 <input name="text" class="input" type="text" required/>
  <button type="submit" value="Submit">Submit</button>
</form>
<div id="todos">
**{{ block "todos" . }}**
 {{ range .todos }}
 <div id="{{.ID}}"> {{.Text}} </div>
 {{ end }}
**{{ end }}**
</div>
```

In the above page, `range` is enclosed in a `block` expression. The `block` section of the page can now be re-rendered by looking it up by name in the template tree. Fir’s alpine.js plugin can enhance this page by preventing the form submission request and sending a json event instead. On receiving the event, Fir server re-renders the `todos` `block` and sends back a html snippet. The fir alpine.js plugin replaces the content of the `todos` `div`  with the returned html snippet.

```go
<form **x-data @submit.prevent="$fir.submit"** name="newTodo" method="post">
 <input name="text" class="input" type="text" required/>
  <button type="submit" value="Submit">Submit</button>
</form>
<div id="todos">
{{ block "todos" . }}
 {{ range .todos }}
 <div id="{{.ID}}"> {{.Text}} </div>
 {{ end }}
{{ end }}
</div>
```

`x-data` attribute on the `newTodo` form declares it as a new Alpine component. `@submit.prevent` is a `x-on` [directive](https://alpinejs.dev/directives/on) which prevents the form submission and calls `$fir.submit` [custom magic function](https://alpinejs.dev/advanced/extending#magic-functions) instead. The `$fir.submit` function reads the Form element data and sends a `fir.Event` over a real-time connection or a regular fetch call if a real-time connection is unavailable. The event is handled on the server and a `patch` operation is sent back. The patch operation is then applied the DOM by the alpine.js plugin.
