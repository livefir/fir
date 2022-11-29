---
title: "Go"
description: ""
lead: ""
date: 2022-11-18T18:23:25+01:00
lastmod: 2022-11-18T18:23:25+01:00
draft: false
images: []
menu:
  docs:
    parent: "api"
    identifier: "go-bf348457b19bdaa1b581f62ae9a5671f"
weight: 998
toc: true
---


## View
### DefaultView
### DefaultErrorView
## Event
## Page
## Patch
### Render
### Patchset
### After
### Append
### Before
### Morph
### Navigate
### Prepend
### Reload
### Remove
### ResetForm
### Store
## Controller
### WithChannelFunc
### WithPubsubAdapter
### WithWebsocketUpgrader
### WithErrorView
### WithEmbedFs
### WithPublicDir
### DisableTemplateCache
### EnableDebugLog
### EnableWatch
### DevelopmentMode
## Subscription
## PubsubAdapter
## Annotations(entgo)
### CreateForm
### UpdateForm
### ListItem
## Template functions
### Sprig functions
### ActiveRoute
### NotActiveRoute
## Error functions
### PatchError
### PatchFormError
### UnsetPatchFormErrors
### PageError
### PageFormError
### ErrInternalServer
### ErrNotFound
### ErrBadRequest
### ErrUnauthorized
### UserError
## Utility functions
### DecodeForm
### DecodeURLValues
### DefaultChannelFunc
### DefaultFuncMap
### GeneratePublicDir
### MinMax
## Globals
### UserIDKey
### DefaultUserErrorMessage
### DefaultViewExtensions
### DefaultWatchExtensions



```go


p1, p2 := func() (error, error) {
    return func(data any)error{

    },
    return func(patch ...Patch)error{

    }
}()

type PageRenderer func(data any) error
type PatchRenderer func(patch ...Patch) error

type TodosPage struct {
  ch chan Event
}

func(t *TodosPage) OnGet(render *PageRenderer, w *http.ResponseWriter, r http.Request) error {
    
      return render(data)
}

func(t *TodosPage) OnPost(render *PageRenderer, w *http.ResponseWriter, r http.Request) error {
    
      return render(data)
}

func(t *TodosPage) OnEvent(render *PatchRenderer, event Event) error {
  return render(
    fir.Morph("#todos", "todo",fir.M{"todos": todo}),
    fir.Morph("#todos", "todo",fir.M{"todos": todo}),
  )
}

opts := fir.PageOption[]{
  fir.EventSender(ch),
  fir.ID("todos"),
  fir.Layout("layout.html"),
  fir.Content("todos.html"),
  fir.LayoutContentName("content"),
  fir.Partials([]string{"todo.html"}),
  fir.Extensions([]string{".html"}),
  fir.Funcs(fir.FuncMap{
    "title": func() string {
      return "Todos"
    },
  }),
}
ch := make(chan Event)
c.Page(&TodosPage{ch: ch}, opts)

{{ fir-block "cities" .}}
<datalist>
{{range .cities}}
  <option value="{{.}}"></option>
{{end}}
</datalist>
{{ end }}

fir.Block("todos",fir.M{"todos": todos})
fir.Template("todos",fir.M{"todos":todos}).Morph("#todos")
fir.Block("todos",fir,M{"todos":toods}).Morph("#todos")
fir.Navigate()
fir.Store()
fir.Remove("#todos")


Morph("#todos").ExecTemplate("todo",fir.M{"todos":todos})
Morph("#todos").Render("todo",fir.M{"todos":todos})
ExecTemplate("todo",fir.M{"todos":todos}).Morph("#todos")
Exec("todos").Data(fir.M{"todos":todos}).Morph("#todos")
Morph("#todos").HTML(fir.ExecTemplate("todos",fir.M{"todos":todos}))
Template("todos").Morph("#todos",fir.M{"todos":todos})
Selector("#todos").Morph().Template("todos",fir.M{"todos":todos}) 
Morph(#todos",Template("todos"),fir.M{"todos":todos})
fir.Morph("#todos",fir.Block("todos",todos))



controller handles routes -> onget, onpost, onevent

view -> layout -> page -> patch

page.Render() -> Renders full page
page.Patch(fir.Morph{
			Selector: "#todos",
			HTML: &fir.Render{
				Template: "todos",
				Data:     map[string]any{"todos": todos},
			},
		}) -> Renders partial page


```

.firErrors.message
.firErrors.status
.firErrors.newTodo.title

fir.Errors(fir.M{"message": "error message", "status": 400, "log": "error log", "newTodo": fir.M{"title": "error title"}})
fmt.Errorf("%v %w",err,fir.M{"status": 400, "newTodo": fir.M{"title": "error title"}})
fmt.Errorf("%v %w",err,fir.M{"title": "error title"}) -> .firErrors.title
err -> .firErrors.message
fir.Morph("#titleError",fir.M{"error": err.Error()})

FormError("title",err)

setError, unsetError := fir.FieldError("title"))
setError(err)
unsetError()
ctx.Patch(setError("title",err))
ctx.FieldError("title",err)

type Form struct {
  Name string `json:"name"`
  Email string `json:"email"`
}

func(f *Form) ValidationErrors()map[string]error{
  return map[string]error{
    "name": errors.New("name is required"),
    "email": errors.New("email is required"),
  }
}


