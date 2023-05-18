package fir

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timshannon/bolthold"
	"github.com/yosssi/gohtml"
	"golang.org/x/net/html"
)

func TestSanity(t *testing.T) {
	dbfile, err := ioutil.TempFile("", "test")
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
	wantPage, err := ioutil.ReadFile("example_golden.html")
	if err != nil {
		t.Fatal(err)
	}
	wantPageNode, err := html.Parse(bytes.NewReader(wantPage))
	if err != nil {
		t.Fatal(err)
	}
	if err := areNodesDeepEqual(gotPageNode, wantPageNode); err != nil {
		t.Fatalf("\nerr: %v \ngot \n %v \n want \n %v", err, gohtml.Format(string(htmlNodeToBytes(gotPageNode))), gohtml.Format(string(htmlNodeToBytes(wantPageNode))))
	}

}
