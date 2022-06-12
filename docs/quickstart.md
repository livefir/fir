---
layout: page
title: Quickstart
permalink: /quickstart/
---

# Quickstart

Lets spend the next 15 minutes creating a new `reactive` counter app.

# Prerequisites

Have you installed [Go](https://go.dev/doc/install) ? If yes, we are good to go.

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
	c := fir.NewController("A counter app", fir.DevelopmentMode(true))
	http.Handle("/", c.Handler(&fir.HelloView{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}

```
