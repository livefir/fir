package fir

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/timshannon/bolthold"
	"github.com/yosssi/gohtml"
	"golang.org/x/net/html"
)

func TestSanity(t *testing.T) {
	dbfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}

	db, err := bolthold.Open(dbfile.Name(), 0666, nil)
	if err != nil {
		panic(err)
	}

	controller := NewController("todos")
	ts := httptest.NewServer(controller.RouteFunc(todosRoute(db)))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	page, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	gotPageNode, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	wantPage, err := os.ReadFile("example_golden.html")
	if err != nil {
		t.Fatal(err)
	}
	wantPageNode, err := html.Parse(bytes.NewReader(wantPage))
	if err != nil {
		t.Fatal(err)
	}
	if err := areNodesDeepEqual(gotPageNode, wantPageNode); err != nil {
		t.Fatalf("\nerr: %v \ngot \n %v \n want \n %v", err,
			gohtml.Format(htmlNodetoString(gotPageNode)),
			gohtml.Format(htmlNodetoString(wantPageNode)))
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	sel := `input[name="text"]`
	todo := "test"
	result := ""
	if err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL),
		chromedp.WaitVisible(sel, chromedp.ByQuery),
		chromedp.SendKeys(sel, todo, chromedp.ByQuery),
		chromedp.Submit(sel, chromedp.ByQuery),
		chromedp.WaitVisible(`div[id="todo-1"]`, chromedp.ByQuery),
		chromedp.TextContent(`div[id="todo-1"]`, &result, chromedp.ByQuery),
	); err != nil {
		t.Fatal(err)
	}
	if removeSpace(result) != todo {
		t.Fatalf("got %q, want %q", result, todo)
	}

}
