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

func (t *TaskView) OnGet(_ http.ResponseWriter, _ *http.Request) fir.Pagedata {
	t.RLock()
	defer t.RUnlock()
	return fir.Pagedata{Data: map[string]any{"tasks": t.tasks}}
}

func (t *TaskView) OnPost(_ http.ResponseWriter, r *http.Request) fir.Pagedata {
	t.Lock()
	defer t.Unlock()

	var task Task
	if err := fir.DecodeForm(&task, r); err != nil {
		return fir.PageError(err, "failed to decode form")
	}

	t.tasks = append(t.tasks, task)
	return fir.Pagedata{Data: map[string]any{"tasks": t.tasks}}
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
				HTML: &fir.Render{
					Template: "tasks",
					Data:     map[string]any{"tasks": t.tasks},
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
