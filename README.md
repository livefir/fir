# Fir

[![Go Reference](https://pkg.go.dev/badge/github.com/livefir/fir.svg)](https://pkg.go.dev/github.com/livefir/fir) 
[![npm version](https://badge.fury.io/js/@livefir%2Ffir.svg)](https://badge.fury.io/js/@livefir%2Ffir)

**A Go toolkit to build reactive web interfaces using: [Go](https://go.dev/), [html/template](https://pkg.go.dev/html/template) and [alpinejs](https://alpinejs.dev/).**

The **Fir** toolkit is designed for Go developers with moderate html/css & js skills who want to progressively build reactive web apps without mastering complex web frameworks. It includes a Go library and an Alpine.js plugin. Fir can be used for building passable UIs with low effort and is a good fit for internal tools, interactive forms and real-time feeds.

Fir has a simple and predictable server and client API with only one new (*and* *weird*) expression:

```jsx
  <div @fir:create-chirp:ok::chirp="$fir.prependEl()">
  ...
  </div>
```

[snippet from chirper example](./examples/chirper/index.html#L45)

Fir’s magic event handling expression `fir:event-name:event-state::template-name` piggybacks on [alpinejs event binding syntax](https://alpinejs.dev/directives/on#custom-events) to declare [html/templates](https://pkg.go.dev/html/template) to be re-rendered on the server. Once accustomed to this weirdness, Fir unlocks a great deal of productivity while adhering to web standards to enable rich and interactive apps

Fir sits somewhere between [phoenix liveview](https://github.com/phoenixframework/phoenix_live_view) and [htmx](https://htmx.org/) in terms of capabilities. It's event-driven like liveview but, instead of providing atomic UI diffs, it returns html fragments like htmx. 

**Feature Highlights:**

- **Server rendered**: Render HTML on the server using Go’s standard html templating library.
- **DOM Patching:** React to user or server events to update only the changed parts of a web page.
- **Publish over websocket**: Broadcast html fragments over websocket in response to both client and server events.
- **Interactivity over standard HTTP**: Fir possesses a built-in pubsub over websocket capability to broadcast UI diff changes to connected clients. However, it doesn't solely rely on websockets. It's still possible to disable websockets and benefit from UI diffs sent over standard HTTP.

## Usage

[Demo & Quickstart](https://livefir.fly.dev/)

[Examples](./examples/)

[How it works](https://adnaan.notion.site/Fir-2358531aced84bf1b0b1a687760fff3b)

## CLI

You don't need this to get started but the the cli can be used to generate the boilerplate:

```go
go run github.com/livefir/fir/cli gen project -n quickstart // generates a folder named quickstart
cd quickstart
go run main.go
open localhost:9867

go run github.com/livefir/fir/cli gen route -n index // generates a new route
```



## Community

- https://blog.logrocket.com/building-reactive-web-app-go-fir/

    - [Building a web application in Go with Fir | Part 1| youtube.com](https://www.youtube.com/watch?v=7hpXdG-Nw00)

- https://bsky.app/profile/hrbrmstr.dev/post/3kk2f5eqxyk22

    - https://dailydrop.hrbrmstr.dev/2024/01/29/drop-410-2024-01-29-your-real-time-app-stack/

