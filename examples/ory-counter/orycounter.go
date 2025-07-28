package orycounter

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/livefir/fir"
	"github.com/livefir/fir/internal/dev"
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
		session, _, err := app.ory.FrontendAPI.ToSession(request.Context()).Cookie(cookies).Execute()
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

type index struct {
	value int32
}

func (i *index) load(ctx fir.RouteContext) error {
	return ctx.Data(map[string]any{"count": atomic.LoadInt32(&i.value)})
}

func (i *index) inc(ctx fir.RouteContext) error {
	return ctx.Data(map[string]any{"count": atomic.AddInt32(&i.value, 1)})
}

func (i *index) dec(ctx fir.RouteContext) error {
	return ctx.Data(map[string]any{"count": atomic.AddInt32(&i.value, -1)})
}

func (i *index) Options() fir.RouteOptions {
	return fir.RouteOptions{
		fir.ID("counter"),
		fir.Content("count.html"),
		fir.OnLoad(i.load),
		fir.OnEvent("inc", i.inc),
		fir.OnEvent("dec", i.dec),
	}
}

func Index() fir.RouteOptions {
	i := &index{}
	return i.Options()
}

func NewRoute() fir.RouteOptions {
	return Index()
}

func Run(port int) error {
	proxyPort := os.Getenv("PROXY_PORT")
	if proxyPort == "" {
		proxyPort = "4000"
	}
	c := ory.NewConfiguration()
	c.Servers = ory.ServerConfigurations{{URL: fmt.Sprintf("http://localhost:%s/.ory", proxyPort)}}

	app := &App{
		ory: ory.NewAPIClient(c),
	}

	dev.SetupAlpinePluginServer()
	controller := fir.NewController("fir-ory-counter", fir.DevelopmentMode(true))
	http.Handle("/", app.sessionMiddleware(controller.Route(&index{})))
	log.Printf("Ory Counter example listening on http://localhost:%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
