package chirper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/livefir/fir"
	"github.com/timshannon/bolthold"
)

func TestChirperIntegration(t *testing.T) {
	// Create a temporary database for testing
	db, err := bolthold.Open("test_chirper.db", 0666, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		db.Close()
		// Clean up test database file
		// Note: In a real test, you might want to use a temp file
	}()

	// Set up the controller and route
	controller := fir.NewController("chirper_test")
	server := httptest.NewServer(controller.RouteFunc(func() fir.RouteOptions {
		return fir.RouteOptions{
			fir.ID("chirper-test"),
			fir.Content("index.html"),
			fir.OnLoad(loadChirps(db)),
			fir.OnEvent("create-chirp", createChirp(db)),
			fir.OnEvent("delete-chirp", deleteChirp(db)),
			fir.OnEvent("like-chirp", likeChirp(db)),
		}
	}))
	defer server.Close()

	// Helper function to get session cookie
	getSessionCookie := func() string {
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		for _, cookie := range resp.Cookies() {
			if cookie.Name == "_fir_session_" {
				return cookie.Value
			}
		}
		return ""
	}

	sessionID := getSessionCookie()
	if sessionID == "" {
		t.Fatal("Could not get session cookie")
	}

	// Test 1: Create a chirp
	t.Run("create_chirp", func(t *testing.T) {
		chirp := Chirp{
			Body: "Test chirp for deletion",
		}

		event := fir.Event{
			ID:        "create-chirp",
			IsForm:    false,
			Params:    mustMarshal(t, chirp),
			SessionID: &sessionID,
			Timestamp: time.Now().UTC().UnixMilli(),
		}

		// Send create event
		payload := mustMarshal(t, event)
		req := httptest.NewRequest("POST", server.URL, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-FIR-MODE", "event")
		req.AddCookie(&http.Cookie{Name: "_fir_session_", Value: sessionID})

		resp := httptest.NewRecorder()
		server.Config.Handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.Code, resp.Body.String())
		}

		// Verify chirp was created
		var chirps []Chirp
		if err := db.Find(&chirps, &bolthold.Query{}); err != nil {
			t.Fatal(err)
		}

		if len(chirps) != 1 {
			t.Fatalf("Expected 1 chirp, got %d", len(chirps))
		}

		if chirps[0].Body != "Test chirp for deletion" {
			t.Fatalf("Expected chirp body 'Test chirp for deletion', got '%s'", chirps[0].Body)
		}
	})

	// Test 2: Delete the chirp
	t.Run("delete_chirp", func(t *testing.T) {
		// Get the chirp ID from database
		var chirps []Chirp
		if err := db.Find(&chirps, &bolthold.Query{}); err != nil {
			t.Fatal(err)
		}

		if len(chirps) != 1 {
			t.Fatalf("Expected 1 chirp before deletion, got %d", len(chirps))
		}

		chirpID := chirps[0].ID

		// Create delete event
		deleteReq := struct {
			ChirpID uint64 `json:"chirpID"`
		}{
			ChirpID: chirpID,
		}

		event := fir.Event{
			ID:        "delete-chirp",
			IsForm:    false,
			Params:    mustMarshal(t, deleteReq),
			SessionID: &sessionID,
			Timestamp: time.Now().UTC().UnixMilli(),
		}

		// Send delete event
		payload := mustMarshal(t, event)
		req := httptest.NewRequest("POST", server.URL, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-FIR-MODE", "event")
		req.AddCookie(&http.Cookie{Name: "_fir_session_", Value: sessionID})

		resp := httptest.NewRecorder()
		server.Config.Handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.Code, resp.Body.String())
		}

		// Verify chirp was deleted from database
		var remainingChirps []Chirp
		if err := db.Find(&remainingChirps, &bolthold.Query{}); err != nil {
			t.Fatal(err)
		}

		if len(remainingChirps) != 0 {
			t.Fatalf("Expected 0 chirps after deletion, got %d", len(remainingChirps))
		}
	})

	// Test 3: Like a chirp
	t.Run("like_chirp", func(t *testing.T) {
		// First create a chirp to like
		chirp := Chirp{
			Body: "Test chirp for liking",
		}

		event := fir.Event{
			ID:        "create-chirp",
			IsForm:    false,
			Params:    mustMarshal(t, chirp),
			SessionID: &sessionID,
			Timestamp: time.Now().UTC().UnixMilli(),
		}

		// Send create event
		payload := mustMarshal(t, event)
		req := httptest.NewRequest("POST", server.URL, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-FIR-MODE", "event")
		req.AddCookie(&http.Cookie{Name: "_fir_session_", Value: sessionID})

		resp := httptest.NewRecorder()
		server.Config.Handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.Code)
		}

		// Get the chirp ID
		var chirps []Chirp
		if err := db.Find(&chirps, &bolthold.Query{}); err != nil {
			t.Fatal(err)
		}

		chirpID := chirps[0].ID
		initialLikes := chirps[0].LikesCount

		// Create like event
		likeReq := struct {
			ChirpID uint64 `json:"chirpID"`
		}{
			ChirpID: chirpID,
		}

		likeEvent := fir.Event{
			ID:        "like-chirp",
			IsForm:    false,
			Params:    mustMarshal(t, likeReq),
			SessionID: &sessionID,
			Timestamp: time.Now().UTC().UnixMilli(),
		}

		// Send like event
		likePayload := mustMarshal(t, likeEvent)
		likeReq2 := httptest.NewRequest("POST", server.URL, bytes.NewReader(likePayload))
		likeReq2.Header.Set("Content-Type", "application/json")
		likeReq2.Header.Set("X-FIR-MODE", "event")
		likeReq2.AddCookie(&http.Cookie{Name: "_fir_session_", Value: sessionID})

		likeResp := httptest.NewRecorder()
		server.Config.Handler.ServeHTTP(likeResp, likeReq2)

		if likeResp.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", likeResp.Code, likeResp.Body.String())
		}

		// Verify like count increased
		var updatedChirps []Chirp
		if err := db.Find(&updatedChirps, &bolthold.Query{}); err != nil {
			t.Fatal(err)
		}

		if len(updatedChirps) != 1 {
			t.Fatalf("Expected 1 chirp, got %d", len(updatedChirps))
		}

		finalLikes := updatedChirps[0].LikesCount
		if finalLikes != initialLikes+1 {
			t.Fatalf("Expected likes to increase from %d to %d, got %d", initialLikes, initialLikes+1, finalLikes)
		}
	})
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
