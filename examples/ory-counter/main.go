package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/adnaan/fir"
	ory "github.com/ory/client-go"
)

type App struct {
	ory *ory.APIClient
	// save the cookies for any upstream calls to the Ory apis
	cookies string
	// save the session to display it on the dashboard
	session *ory.Session
}

func (app *App) sessionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		log.Printf("handling middleware request\n")

		// set the cookies on the ory client
		// this example passes all request.Cookies
		// to `ToSession` function
		//
		// However, you can pass only the value of
		// ory_session_projectid cookie to the endpoint
		cookies := request.Header.Get("Cookie")

		// check if we have a session
		session, _, err := app.ory.V0alpha2Api.ToSession(request.Context()).Cookie(cookies).Execute()
		if (err != nil && session == nil) || (err == nil && !*session.Active) {
			// this will redirect the user to the managed Ory Login UI
			http.Redirect(writer, request, "/.ory/self-service/login/browser", http.StatusSeeOther)
			return
		}
		app.cookies = cookies
		app.session = session
		// continue to the requested page (in our case the Dashboard)
		next.ServeHTTP(writer, request)
	}
}

type Counter struct {
	count int32
}

func morphCount(c int32) fir.Patch {
	return fir.Morph{
		Selector: "#count",
		Template: &fir.Template{
			Name: "count",
			Data: fir.Data{"count": c},
		},
	}
}

func (c *Counter) Inc() fir.Patch {
	return morphCount(atomic.AddInt32(&c.count, 1))
}

func (c *Counter) Dec() fir.Patch {
	return morphCount(atomic.AddInt32(&c.count, -1))
}

func (c *Counter) Value() int32 {
	return atomic.LoadInt32(&c.count)
}

type CounterView struct {
	fir.DefaultView
	model *Counter
}

func (c *CounterView) Content() string {
	return `<!DOCTYPE html>
	<html lang="en">
	
	<head>
		<title>{{.app_name}}</title>
		<meta charset="UTF-8">
		<meta name="description" content="A counter app">
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css" />
		<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/fir.min.js"></script>
		<script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
	</head>

	<body>
		<div class="my-6" style="height: 500px">
			<div class="columns is-mobile is-centered is-vcentered">
				<div x-data class="column is-one-third-desktop has-text-centered is-narrow">
					<div>
						{{block "count" .}}<div id="count">{{.count}}</div>{{end}}
						<button class="button has-background-primary" @click="$fir.emit('inc')">+
						</button>
						<button class="button has-background-primary" @click="$fir.emit('dec')">-
						</button>
					</div>
				</div>
			</div>
		</div>
	</body>
	
	</html>`
}

func (c *CounterView) OnGet(_ http.ResponseWriter, _ *http.Request) fir.Page {
	return fir.Page{
		Data: fir.Data{
			"count": c.model.Value(),
		}}
}

func (c *CounterView) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "inc":
		return fir.Patchset{c.model.Inc()}
	case "dec":
		return fir.Patchset{c.model.Dec()}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}

	return nil
}

func main() {
	proxyPort := os.Getenv("PROXY_PORT")
	if proxyPort == "" {
		proxyPort = "4000"
	}
	c := ory.NewConfiguration()
	c.Servers = ory.ServerConfigurations{{URL: fmt.Sprintf("http://localhost:%s/.ory", proxyPort)}}

	app := &App{
		ory: ory.NewAPIClient(c),
	}

	controller := fir.NewController("fir-ory-counter", fir.DevelopmentMode(true))
	http.Handle("/", app.sessionMiddleware(controller.Handler(&CounterView{model: &Counter{}})))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
