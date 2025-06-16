
# Chirper: A Simple Real-Time Twitter Clone

The example demonstrates **progressive enhancement**, **form validation**, and **real-time changes**.

---

## Table of Contents
- [Prerequisites](#prerequisites)
- [Step 1: Setting Up the Project](#step-1-setting-up-the-project)
- [Step 2: Features Overview](#step-2-features-overview)
  - [Progressive Enhancement](#1-progressive-enhancement)
  - [Listing Chirps](#2-listing-chirps)
  - [Creating Chirps](#3-creating-chirps)
  - [Liking and Deleting Chirps](#4-liking-and-deleting-chirps)
  - [Error Handling](#5-error-handling)
  - [Real-Time Updates with WebSocket](#6-real-time-updates-with-websocket)
- [Step 3: Testing the Application](#step-3-testing-the-application)
- [Conclusion](#conclusion)
- [References](#references)

---

## Prerequisites

Before starting, ensure you have the following installed:

- [Go](https://go.dev/) (1.18 or later)
- [Bolthold](https://github.com/timshannon/bolthold) (used for database storage)
- Basic knowledge of Go, HTML, and JavaScript

---

## Step 1: Setting Up the Project

1. Clone the Fir repository:

   ```bash
   git clone https://github.com/livefir/fir.git
   ```

2. Navigate to the Chirper example directory:

   ```bash
   cd fir/examples/chirper
   ```

3. Ensure you have the required dependencies installed:
   - Install [Bolthold](https://github.com/timshannon/bolthold) for database storage.
   - Install Go (1.18 or later).

4. Run the application:

   ```bash
   go run .
   ```

5. Open your browser and navigate to:
   - [http://localhost:9867/nojs](http://localhost:9867/nojs) for the plain HTML version.
   - [http://localhost:9867](http://localhost:9867) for the enhanced JavaScript version.

---

## Step 2: Features Overview

### 1. Progressive Enhancement

Chirper supports both a plain HTML version (for browsers without JavaScript) and an enhanced version with JavaScript for real-time updates.

#### Plain HTML Version ([`index_no_js.html`](./index_no_js.html))

- **Route**: Handled by the [`NoJSIndex`](./index.go) function in [`index.go`](./index.go).
- **How it works**:
  - The `OnLoad` function (`loadChirps`) is invoked automatically when the page loads.
  - Chirps are fetched from the database and rendered using the `chirps` data.
  - Errors during loading are displayed using `{{ fir.Error "onload" }}`.

#### Enhanced Version ([`index.html`](./index.html))

- **Route**: Handled by the [`Index`](./index.go) function in [`index.go`](./index.go).
- **How it works**:
  - Uses Alpine.js and Fir's JavaScript plugin to enable real-time updates.
  - Events like creating, liking, and deleting chirps are handled without reloading the page.

---

### 2. Listing Chirps

- Chirps are fetched from the database using the `loadChirps` function in [`index.go`](./index.go).
- The returned data (`ctx.Data`) is used to render the page.
- Errors during loading are displayed using `{{ fir.Error "onload" }}`.

#### Plain HTML Example:

```html
<p>
    {{ fir.Error "onload" }}
</p>
<div>
    {{ range .chirps }}
        <section>
            <blockquote>{{ .Body }}</blockquote>
            <footer>
                <button formaction="?event=like-chirp" type="submit">&#9829; {{ .LikesCount }}</button>
                <button formaction="?event=delete-chirp" type="submit">&#10005;</button>
            </footer>
        </section>
    {{ end }}
</div>
```

---

### 3. Creating Chirps

#### Plain HTML Version

- Chirps are created by submitting a form with the action `?event=create-chirp`.
- The `createChirp` function in [`index.go`](./index.go):
  - Validates the chirp body (minimum 3 characters).
  - Uses `RouteContext.Bind` to bind form data and return errors.
  - Errors are displayed using `{{ fir.Error "create-chirp.body" }}`.

Example:

```html
<form method="post" action="?event=create-chirp">
    <textarea name="body" placeholder="a new chirp" rows="4" cols="100"></textarea>
    <p>{{ fir.Error "create-chirp.body" }}</p>
    <button type="submit">Chirp</button>
</form>
```

#### Enhanced Version

- Prevents page reloads using Alpine.js and Fir's `$fir.submit()` magic function.
- Resets the form on success using `@fir:create-chirp:ok="$el.reset()"`.
- Example:

```html
<form
    method="post"
    action="?event=create-chirp"
    @submit.prevent="$fir.submit()"
    @fir:create-chirp:ok="$el.reset()">
    <textarea name="body" placeholder="a new chirp" rows="4" cols="100"></textarea>
    <p @fir:create-chirp:error="$fir.replace()">{{ fir.Error "create-chirp.body" }}</p>
    <button type="submit">Chirp</button>
</form>
```

---

### 4. Liking and Deleting Chirps

#### Plain HTML Version

- Chirps can be liked or deleted using forms with actions `?event=like-chirp` and `?event=delete-chirp`.
- Example:

```html
<form method="post">
    <button formaction="?event=like-chirp" type="submit">&#9829; {{ .LikesCount }}</button>
    <button formaction="?event=delete-chirp" type="submit">&#10005;</button>
</form>
```

#### Enhanced Version

- Real-time updates are enabled using Alpine.js and Fir's magic functions:
  - `@fir:like-chirp:ok="$fir.replace()"` updates the like count.
  - `@fir:delete-chirp:ok="$fir.removeEl()"` removes the chirp.

Example:

```html
<section fir-key="{{ .ID }}" @fir:delete-chirp:ok="$fir.removeEl()">
    <form method="post" @submit.prevent="$fir.submit()">
        <blockquote>
            {{ .Body }}
        </blockquote>
        <input type="hidden" name="chirpID" value="{{ .ID }}" />
        <footer>
            <button @fir:like-chirp:ok="$fir.replace()" formaction="?event=like-chirp" type="submit">&#9829; {{ .LikesCount }}</button>
            <button formaction="?event=delete-chirp" type="submit">&#10005;</button>
        </footer>
    </form>
</section>
```

---

### 5. Error Handling

- Errors are displayed using the `fir.Error` template function.
- Examples:
  - Global errors: `{{ fir.Error "onload" }}`
  - Field-specific errors: `{{ fir.Error "create-chirp.body" }}`

---

### 6. Real-Time Updates with WebSocket

- The enhanced version uses WebSocket for real-time updates.
- If WebSocket is disabled, events are sent as HTTP POST requests.

---

## Step 3: Testing the Application

1. Open two browser tabs at [http://localhost:9867](http://localhost:9867).
2. Create a chirp in one tab and see it appear instantly in the other tab.
3. Like or delete chirps and observe real-time updates.

---

## Conclusion

Congratulations! You've built a simple real-time Twitter clone using Fir. This example demonstrates how Fir enables progressive enhancement, form validation, and real-time updates with minimal effort. Explore the code further to customize and extend Chirper!

## References
- [Fir Documentation](https://github.com/livefir/fir)
- [Alpine.js Documentation](https://alpinejs.dev/)
- [Bolthold Documentation](https://github.com/timshannon/bolthold)