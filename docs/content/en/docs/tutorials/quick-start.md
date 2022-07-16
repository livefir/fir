---
title: "Quick Start"
description: "A quick start for the Fir toolkit"
lead: "A quick start for the Fir toolkit"
date: 2020-11-16T13:59:39+01:00
lastmod: 2020-11-16T13:59:39+01:00
draft: false
images: []
menu:
  docs:
    parent: "tutorials"
weight: 110
toc: true
---

Lets spend the next 15 minutes creating a simple app. If you want to skip ahead and look at final code, its here: [examples/tasks/main.go](https://github.com/adnaan/fir/blob/main/examples/tasks/main.go)

## Prerequisites

Have you installed [Go 1.18](https://go.dev/doc/install) ? If yes, we are good to go.


## Creating a new app


The `fir` library concerns itself with only the view controller so starting off is as easy as mounting a view on the `fir` controller:

```go
package main

import (
 "log"
 "net/http"

 "github.com/adnaan/fir"
)

func main() {
 controller := fir.NewController("task_app", fir.DevelopmentMode(true))
 http.Handle("/", controller.Handler(&fir.DefaultView{}))
 http.ListenAndServe(":9867", nil)
}

```

```bash
mkdir hello-fir && cd hello-fir
touch main.go # copy-paste the above snippet
go get ./...
go run main.go
```

Open [localhost:9867](http://localhost:9867) to see the running app.

We have created a controller and registered a `DefaultView` by calling `controller.Handler(&fir.HelloView{})`. The `contoller.Handler` method accepts a [View](https://pkg.go.dev/github.com/adnaan/fir#View) interface. `fir.DefaultView` satisfies the methods for the `View` interface with default values.

The fir library doesn't manage routing so you can bring your favorite routing library to actually route requests to the view. Here we keep it simple and mount the `http.HandlerFunc` returned by `controller.Handler` on the `/` route: `http.Handle("/", c.Handler(&fir.DefaultView{}))`

## Creating a new view

Lets start rendering a simple html page. To do this we want to create a new view and replace `DefaultView`.

This is how we do that:

```go
type TaskView struct {
 fir.DefaultView
}

func (*TaskView) Content() string {
 return "A tasks app"
}

```

In the above snippet we have created a new struct, `TaskView` and embedded a `fir.DefaultView` type in it to satisfy the `View` interface.

```go
package main

import (
 "log"
 "net/http"

 "github.com/adnaan/fir"
)

type Task struct {
 Text string `json:"text" schema:"text"`
}

type TaskView struct {
 fir.DefaultView
 tasks []Task
 sync.RWMutex
}

func (*TaskView) Content() string {
 return "A tasks app"
}

func main() {
 controller := fir.NewController("task_app", fir.DevelopmentMode(true))
 http.Handle("/", controller.Handler(&TaskView{tasks: make([]Task, 0)}))
 http.ListenAndServe(":9867", nil)
}
```

Run the above code to see the changes at [localhost:9867](http://localhost:9867).

## Render Page

A `Page` in `fir` is a full web page which is rendered when a route is loaded. The `View` interface exposes two methods to render pages: `OnGet` and `OnPost`. `OnGet` is called when the page is loaded with a http GET request. We override `OnGet` to supply data to `html/template` renderer. `OnPost` can be overriden to handle form submissions from the browser. Both of these methods reload and re-render the full page.

```go
package main

import (
 "net/http"
 "sync"

 "github.com/adnaan/fir"
)

type Task struct {
 Text string `json:"text" schema:"text"`
}

type TaskView struct {
 fir.DefaultView
 tasks []Task
 sync.RWMutex
}

func (t *TaskView) OnGet(_ http.ResponseWriter, _ *http.Request) fir.Page {
 t.RLock()
 defer t.RUnlock()
 return fir.Page{Data: fir.Data{"tasks": t.tasks}}
}

func (t *TaskView) OnPost(_ http.ResponseWriter, r *http.Request) fir.Page {
 t.Lock()
 defer t.Unlock()

 var task Task
 if err := fir.DecodeForm(&task, r); err != nil {
  return fir.PageError(err, "failed to decode form")
 }

 t.tasks = append(t.tasks, task)
 return fir.Page{Data: fir.Data{"tasks": t.tasks}}
}

func (*TaskView) Content() string {
 return `
 <div>
  <h1>Tasks</h1>
  <form id="new-task" method="post">
   <input type="text" name="text" placeholder="New task" />
  </form>
  {{range .tasks}}
   <div>{{.Text}}</div>
  {{end}}
 </div>`
}

func main() {
 controller := fir.NewController("task_app", fir.DevelopmentMode(true))
 http.Handle("/", controller.Handler(&TaskView{tasks: make([]Task, 0)}))
 http.ListenAndServe(":9867", nil)
}
```

Run the above update code. Go to [localhost](http://localhost:9867) submit a new task.

Creating a new task is handled by `OnPost` which appends a new task to `tasks` and then re-renders the `Page` with updated `tasks`.

## Patch Page

In the above example, we reload the entire page when a new task is created. `Fir` offers a way to re-render sections of a page without reloading the page. We call this a `Patch` operation. If we look at the html being rendered, we can see that only following snippet needs to be re-rendered instead of the whole page:

```html
{{range .tasks}}
 <div>{{.Text}}</div>
{{end}}
```

Since we want to re-render this section of our html on the server, we will wrap it around in a defined template(html/template). We will use the [block](https://pkg.go.dev/text/template#example-Template-Block) action.

The updated html snippet should look like this:

```gohtml
{{block "tasks" .}}
 <div id="tasks">
  {{range .tasks}}
   <div>{{.Text}}</div>
  {{end}}
 </div>
{{end}}
```

The  `range` action is wrapped in a `<div id="tasks">` so that we can target that html section for re-rendering it. We now need a way to handle the form submission and respond with a `Patch` operation to update the changed section.

### Emit page events

`Fir` has a companion javascript library which lets you send browser events to the server. You can use these events to change server state(in our case: `tasks []Task`) and make partial page updates without a page reload. Lets include the `fir` javascript library in our html page by overriding the `Layout` method of the `View` interface. We have also updated the `Content()` method:

1. A defined template named `content`(`{{define "content"}} ... {{end}}{%end%}`) which can be replaced in the `Layout` html string.
2. `<div x-data>` so that we can use alpinejs features within this div tag.

```go
func (*TaskView) Layout() string {
 return `
 <!DOCTYPE html>
 <html lang="en">
 <head>
  <title>{{.app_name}}</title>
  <script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/fir.min.js"></script>
  <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
 </head>
 <body>
  {{template "content" .}}
 </body>
 </html>`
}

func (*TaskView) Content() string {
 return `
 {{define "content"}}
  <div x-data>
   <h1>Tasks</h1>
   <form id="new-task" method="post" @submit.prevent="$fir.submit">
    <input type="text" name="text" placeholder="New task" />
   </form>
   {{block "tasks" .}}
    <div id="tasks">
     {{range .tasks}}
      <div>{{.Text}}</div>
     {{end}}
    </div>
   {{end}}
  </div>
 {{end}}`
}
```

Now we can use [alpinejs](https://alpinejs.dev/directives/on#prevent) `@submit.prevent` binding to call a utility function from the library: `$fir.submit`.

```html
<form id="new-task" method="post" @submit.prevent="$fir.submit">
 <input type="text" name="text" placeholder="New task" />
</form>
```

In the above snippet, we use the custom Alpinejs magic function, `$fir.submit` to send an event to the server on form submission. Internally `$fir.submit` collects the form data and sends a `post` request to the controller.  Shortly we will see how to handle this event to change state on the server, followed by updating tasks on the web page.

### Handle page events

To handle page events emitted by the `fir` client library, we override the `OnEvent` method of the `View` interface. In response to a page event, we want to send back a set of `Patch` operations which can modify targeted sections of the page.

```go
...

func (t *TaskView) OnEvent(event fir.Event) fir.Patchset {
 switch event.ID {
 case "new-task":
  var task Task
  if err := event.DecodeFormParams(&task); err != nil {
   return fir.PatchError(err, "failed to decode task")
  }

  t.Lock()
  defer t.Unlock()
  t.tasks = append(t.tasks, task)
  return fir.Patchset{
   fir.Morph{
    Selector: "#tasks",
    Template: &fir.Template{
     Name: "tasks",
     Data: fir.Data{"tasks": t.tasks},
    },
   },
  }
 }
 return nil
}
```

## Final

The big idea behind `Fir` is wrapping a `div` in a `{{ template ...}}` `html/template` action and then later patching it over standard HTTP on state change. Below is the final code:

```go
package main

import (
	"net/http"
	"sync"

	"github.com/adnaan/fir"
)

type Task struct {
	Text string `json:"text" schema:"text"`
}

type TaskView struct {
	fir.DefaultView
	tasks []Task
	sync.RWMutex
}

func (t *TaskView) OnGet(_ http.ResponseWriter, _ *http.Request) fir.Page {
	t.RLock()
	defer t.RUnlock()
	return fir.Page{Data: fir.Data{"tasks": t.tasks}}
}

func (t *TaskView) OnPost(_ http.ResponseWriter, r *http.Request) fir.Page {
	t.Lock()
	defer t.Unlock()

	var task Task
	if err := fir.DecodeForm(&task, r); err != nil {
		return fir.PageError(err, "failed to decode form")
	}

	t.tasks = append(t.tasks, task)
	return fir.Page{Data: fir.Data{"tasks": t.tasks}}
}

func (t *TaskView) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "new-task":
		var task Task
		if err := event.DecodeFormParams(&task); err != nil {
			return fir.PatchError(err, "failed to decode task")
		}

		t.Lock()
		defer t.Unlock()
		t.tasks = append(t.tasks, task)
		return fir.Patchset{
			fir.Morph{
				Selector: "#tasks",
				Template: &fir.Template{
					Name: "tasks",
					Data: fir.Data{"tasks": t.tasks},
				},
			},
		}
	}
	return nil
}

func (*TaskView) Layout() string {
	return `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<title>{{.app_name}}</title>
		<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/fir.min.js"></script>
		<script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
	</head>
	<body>
		{{template "content" .}}
	</body>
	</html>`
}

func (*TaskView) Content() string {
	return `
	{{define "content"}}
		<div x-data>
			<h1>Tasks</h1>
			<form id="new-task" method="post" @submit.prevent="$fir.submit">
				<input type="text" name="text" placeholder="New task" />
			</form>
			{{block "tasks" .}}
				<div id="tasks">
					{{range .tasks}}
						<div>{{.Text}}</div>
					{{end}}
				</div>
			{{end}}
		</div>
	{{end}}`
}

func main() {
	controller := fir.NewController("task_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(&TaskView{tasks: make([]Task, 0)}))
	http.ListenAndServe(":9867", nil)
}


```


