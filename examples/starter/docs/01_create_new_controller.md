# Create a new controller

The fir controller renders html and manages websocket connections. It's possible to create multiple fir
controllers. Each controller can manage a collections of views for a specific functionality(e.g. dashboard, accounts).
Having multiple controllers can help large apps to separate busy areas from non-busy areas. The separation means that
one can dedicate more resources to a specific area.

For the purposes of this walkthrough, we will focus on a single controller.

Creating a fir controller is easy:

```go
c := fir.Websocket("fir-starter", fir.DevelopmentMode(mode))
```

Were fir-starter is the controller name and the second argument takes functional options. For a full list please
see: https://pkg.go.dev/github.com/adnaan/fir/controller#Option

The fir controller has a single exposed API: [Handler(view View) http.HandlerFunc](https://pkg.go.dev/github.com/adnaan/fir/controller#Controller)
Since the `Handler` api returns [http#HandlerFunc](https://pkg.go.dev/net/http#HandlerFunc), it can be mounted on your
favorite router/muxer.

The `Handler` api accepts a [View](https://pkg.go.dev/github.com/adnaan/fir/controller#View) interface. The `View` type
represents a html page to be rendered.

Finally,

```go
r := chi.NewRouter()
...
r.NotFound(c.Handler(&views.NotfoundView{}))
```

Where `views.NotfoundView` is a struct which implements the [View](https://pkg.go.dev/github.com/adnaan/fir/controller#View) interface.
