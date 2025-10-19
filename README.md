# Fir

[![Go API Reference](https://pkg.go.dev/badge/github.com/livefir/fir.svg)](https://pkg.go.dev/github.com/livefir/fir)
[![Alpine API Reference](https://img.shields.io/badge/alpine_plugin-reference-blue)](./alpinejs-plugin/README.md) 
[![npm version](https://badge.fury.io/js/@livefir%2Ffir.svg)](https://badge.fury.io/js/@livefir%2Ffir)

**A Go toolkit to build reactive web interfaces using: [Go](https://go.dev/), [html/template](https://pkg.go.dev/html/template) and [alpinejs](https://alpinejs.dev/).**

The **Fir** toolkit is designed for **Go** developers with moderate html/css & js skills who want to progressively build reactive web apps without mastering complex web frameworks. It includes a Go library and an Alpine.js plugin. Fir can be used for building passable UIs with low effort and is a good fit for internal tools, interactive forms and real-time feeds.

Fir has a simple and predictable server and client API with only one new (*and* *weird*) expression:

```jsx
  <div @fir:create-chirp:ok::chirp="$fir.prependEl()">
  ...
  </div>
```

[snippet from chirper example](./examples/chirper/index.html#L45)

Fir’s magic event handling expression `fir:event-name:event-state::template-name` piggybacks on [alpinejs event binding syntax](https://alpinejs.dev/directives/on#custom-events) to declare [html/templates](https://pkg.go.dev/html/template) to be re-rendered on the server. Once accustomed to this weirdness, Fir unlocks a great deal of productivity while adhering to web standards to enable rich and interactive apps

Fir sits somewhere between [Phoenix Liveview](https://github.com/phoenixframework/phoenix_live_view) and [htmx](https://htmx.org/) in terms of capabilities. It's event-driven like Liveview but, instead of providing atomic UI diffs, it returns html fragments like htmx. Fir is closer to liveview since it ships with a server library. Similar to Liveview, Fir has the concept of routes: [Fir Route](https://pkg.go.dev/github.com/livefir/fir@main#Route), [Liveview Router](https://hexdocs.pm/phoenix_live_view/Phoenix.LiveView.Router.html), loading data on page load: [Fir OnLoad](https://pkg.go.dev/github.com/livefir/fir@main#OnLoad), [Liveview onmount](https://hexdocs.pm/phoenix_live_view/Phoenix.LiveView.html#on_mount/1) and handling user events: [Fir OnEvent](https://pkg.go.dev/github.com/livefir/fir@main#OnEvent), [Liveview handle_event](https://hexdocs.pm/phoenix_live_view/Phoenix.LiveView.html#c:handle_event/3).

**Feature Highlights:**

- **Server rendered**: Render HTML on the server using Go’s standard html templating library.
- **DOM Patching:** React to user or server events to update only the changed parts of a web page.
- **Progressive enhancement**: Begin with a JavaScript-free HTML file and Fir's Go server API for quick setup. Gradually improve to avoid page reloads, using Fir's Alpine.js plugin for DOM updates.
- **Publish over websocket**: Broadcast html fragments over websocket in response to both client and server events.
- **Interactivity over standard HTTP**: Fir possesses a built-in pubsub over websocket capability to broadcast UI diff changes to connected clients. However, it doesn't solely rely on websockets. It's still possible to disable websockets and benefit from UI diffs sent over standard HTTP.
- **Broadcast from server**: Broadcast page changes to specific connected clients.
- **Error tracking**: Show and hide user specific errors on the page by simply returning an error or nil.
- **Development live reload**: HTML pages reload automatically on edits if development mode is enabled


## Usage

[Examples](./examples/)

[How it works](https://adnaan.notion.site/Fir-2358531aced84bf1b0b1a687760fff3b)

## CLI

You don't need this to get started but the the cli can be used to generate a simple quickstart boilerplate:

```go
go run github.com/livefir/cli gen project -n quickstart // generates a folder named quickstart
cd quickstart
go run main.go
open localhost:9867

go run github.com/livefir/cli gen route -n index // generates a new route
```



## Community

- https://blog.logrocket.com/building-reactive-web-app-go-fir/

    - [Building a web application in Go with Fir | Part 1| youtube.com](https://www.youtube.com/watch?v=7hpXdG-Nw00)

- https://bsky.app/profile/hrbrmstr.dev/post/3kk2f5eqxyk22

    - https://dailydrop.hrbrmstr.dev/2024/01/29/drop-410-2024-01-29-your-real-time-app-stack/

