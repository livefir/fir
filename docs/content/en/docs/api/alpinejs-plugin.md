---
title: "Alpine.js Plugin"
description: ""
lead: ""
date: 2022-11-18T18:23:35+01:00
lastmod: 2022-11-18T18:23:35+01:00
draft: false
images: []
menu:
  docs:
    parent: "api"
    identifier: "alpinejs-plugin-5ead2ac9ac5f3553b20248acf99f8275"
weight: 999
toc: true
---


## $fir.emit(id, params)

`$fir.emit` is a magic function that returns an event handler(`function(event){...}`) that emits an event with the given id and params to the current `fir` route handler.

`id` must be a string and is required. If not passed as an argument, `id` is picked from the DOM in the following order:

1. element.id
2. form.action="/?event=myevent"
3. button.formaction="/?event=myevent"

Usage:

```html
<button @click="$fir.emit('inc')">+</button>
<button id="dec"@click="$fir.emit()">-</button>
```

```html
    <form
        id="createTodo"
        method="post"
        @submit.prevent="$fir.emit()">
        <input type="text" name="todo" placeholder="a new todo" />
        <button type="submit">Submit</button>
    </form>
```

`params` can be of any type and is optional. If not passed as an argument:

1. `params` composed as an object with event.target.name as key and event.target.value as value.
2. `params` is composed from FormData of the form element if event.target is a form element.

### Events

Browser events are dispatched on start and end of `$fir.emit(...)` call.

Usage:

```html
<div x-data="{loading: false}" 
    @fir:emit-start:myevent.window="loading = true" 
    @fir:emit-end:myevent.window="loading = false">
    <button x-on:click="$fir.emit('myevent')">Click me</button>
    <div x-show="loading">Loading...</div>
</div>
```

### fir:emit

Emitted when `$fir.emit(...)` call is started and finished. It can be used to toggle a loading indicator.

Usage:

```html
<div x-data="{loading: false}" @ir:emit:myevent.window="loading = !loading">
    <button x-on:click="$fir.emit('myevent')">Click me</button>
    <div x-show="loading">Loading...</div>
</div>
```

### fir:emit-start

Emitted when `$fir.emit(...)` call is started.

Usage:

```html
<div @fir:emit-start:myevent.window="loading = true" ></div>
```

### fir:emit-end

Emitted when `$fir.emit(...)` call is finished.

Usage:

```html
<div @fir:emit-end:myevent.window="loading = false" ></div>
```
