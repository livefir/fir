package sanity

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"unicode"

	"github.com/chromedp/chromedp"
	"github.com/livefir/fir"
	"github.com/timshannon/bolthold"
)

func TestSanity(t *testing.T) {
	dbfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dbfile.Name())

	db, err := bolthold.Open(dbfile.Name(), 0666, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	controller := fir.NewController("todos")

	ts := httptest.NewServer(controller.RouteFunc(todosRoute(db)))
	defer ts.Close()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	inputSel := `input[name="text"]`
	submitSel := `button[type="submit"]`
	todo := "test todo"
	var result string

	if err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL),
		chromedp.WaitVisible(inputSel, chromedp.ByQuery),
		chromedp.SendKeys(inputSel, todo, chromedp.ByQuery),
		chromedp.Click(submitSel, chromedp.ByQuery),
		chromedp.WaitVisible(`#todo-1`, chromedp.ByQuery),
		chromedp.TextContent(`#todo-1`, &result, chromedp.ByQuery),
	); err != nil {
		t.Fatal(err)
	}

	if removeSpace(result) != removeSpace(todo) {
		t.Fatalf("got %q, want %q", result, todo)
	}
}

func removeSpace(s string) string {
	rr := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			rr = append(rr, r)
		}
	}
	return string(rr)
}

func TestSanityRaceCondition(t *testing.T) {
	dbfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dbfile.Name())

	db, err := bolthold.Open(dbfile.Name(), 0666, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	controller := fir.NewController("todos")

	ts := httptest.NewServer(controller.RouteFunc(todosRoute(db)))
	defer ts.Close()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	inputSel := `input[name="text"]`
	submitSel := `form#create button[type="submit"]`

	// Test rapid creation of multiple todos to test race conditions
	todos := []string{"todo1", "todo2", "todo3", "todo4", "todo5"}

	var actions []chromedp.Action
	actions = append(actions,
		chromedp.Navigate(ts.URL),
		chromedp.WaitVisible(inputSel, chromedp.ByQuery),
	)

	// Rapidly create multiple todos
	for i, todo := range todos {
		actions = append(actions,
			chromedp.Clear(inputSel, chromedp.ByQuery),
			chromedp.SendKeys(inputSel, todo, chromedp.ByQuery),
			chromedp.Click(submitSel, chromedp.ByQuery),
			// Wait for the todo to appear
			chromedp.WaitVisible(fmt.Sprintf(`#todo-%d`, i+1), chromedp.ByQuery),
			// Small delay to simulate real user interaction timing
			chromedp.Sleep(100*time.Millisecond),
		)
	}

	// Now rapidly toggle done status on multiple todos to test concurrent state updates
	for i := 1; i <= len(todos); i++ {
		toggleButtonSel := fmt.Sprintf(`div[fir-key="%d"] button[formaction*="toggle-done"]`, i)
		actions = append(actions,
			chromedp.Click(toggleButtonSel, chromedp.ByQuery),
			chromedp.Sleep(50*time.Millisecond), // Very short delay to test race conditions
		)
	}

	// Wait a bit for all operations to complete
	actions = append(actions, chromedp.Sleep(500*time.Millisecond))

	// Verify all todos exist and have correct state
	for i, expectedText := range todos {
		todoID := fmt.Sprintf(`#todo-%d`, i+1)
		var actualText string
		actions = append(actions,
			chromedp.TextContent(todoID, &actualText, chromedp.ByQuery),
			chromedp.ActionFunc(func(ctx context.Context) error {
				if removeSpace(actualText) != removeSpace(expectedText) {
					return fmt.Errorf("todo %d: expected %q, got %q", i+1, expectedText, actualText)
				}
				t.Logf("Todo %d verified: %q", i+1, actualText)
				return nil
			}),
		)
	}

	if err := chromedp.Run(ctx, actions...); err != nil {
		t.Fatal(err)
	}

	// Verify database consistency - check that all todos were actually persisted
	var dbTodos []Todo
	if err := db.Find(&dbTodos, &bolthold.Query{}); err != nil {
		t.Fatal(err)
	}

	if len(dbTodos) != len(todos) {
		t.Fatalf("Expected %d todos in database, got %d", len(todos), len(dbTodos))
	}

	// Verify all todos have Done=true (since we toggled them all)
	for i, dbTodo := range dbTodos {
		if !dbTodo.Done {
			t.Errorf("Todo %d should be done but isn't: %+v", i+1, dbTodo)
		}
		expectedText := todos[i]
		if dbTodo.Text != expectedText {
			t.Errorf("Todo %d text mismatch: expected %q, got %q", i+1, expectedText, dbTodo.Text)
		}
	}

	t.Logf("Race condition test passed: created %d todos rapidly and toggled all states successfully", len(todos))
}
