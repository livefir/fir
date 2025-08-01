# Introduction

The `@livefir/fir` package is a companion [alpine.js](https://alpinejs.dev/advanced/extending#via-script-tag) plugin for the Go library [pkg.go.dev/github.com/livefir/fir](https://pkg.go.dev/github.com/livefir/fir). Fir is a Go toolkit to build reactive web interfaces using: [Go](https://go.dev/), [html/template](https://pkg.go.dev/html/template) and [alpinejs](https://alpinejs.dev/). 

## Installation

The plugin can be included as a script tag before the `alpine.js` include.

```html
    <head>
        <script
            defer
            src="https://unpkg.com/@livefir/fir@latest/dist/fir.min.js"></script>
        <script
            defer
            src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
    </head>
```


 See [example](https://github.com/livefir/fir#example).


 ## API

 The plugin provides event handlers [custom magic and directives](https://alpinejs.dev/advanced/extending).


 ### Event handlers

 Fir depends on alpinejs custom event handlers to update the DOM. Event handlers are of the syntax:

 `<x-on:@>:fir:event-id:event-state<ok|error|pending|done>::template-name<optional>`

 For e.g,

 ```html
 <div @fir:create-todo:ok::todo="$fir.prependEl()">
    {{ range .todos }}
        {{ block "todo" . }}

        {{end}}
    {{end}}
</div>
 ```

 The expression `@fir:create-todo:ok::todo` means that if `create-todo`event handler returns with a non-error response then re-render the template named `todo`and send the html to the browser. The event handler utility function `$fir.prependEl()`picks the returned html from `event.detail.html`and prepends it to the current element. For a full list of utiliy functions see section [Magics](#magics)


The `::template-name` binding is optional. It is useful for situations where one needs to re-use a template out-of-band like the example above.

Fir automatically extracts a template from any element which contains html/template actions.

For e.g.

```html
<p @fir:create-todo:error="$fir.replace()">
    {{ fir.Error "create-todo.text" }}
</p>
```

Here Fir automatically extracts a template out of the content of `p`tag. Under the hood, this is how it looks like:

```html
<p @fir:create-todo:error::fir-kxkccoas8d9="$fir.replace()">
    {{ fir.Error "create-todo.text" }}
</p>
```

#### Event states

- ok: when [OnEvent](https://pkg.go.dev/github.com/livefir/fir@main#OnEvent) handler returns a non-error response.
- error: when OnEvent returns an error response.
- pending: client-only for loader states. triggered before the Event is sent to the server.
- done: client-only for loader states. triggered on both ok and error response from the server.

#### Transport Layer

The Fir plugin supports two transport mechanisms with automatic fallback:

- **WebSocket Mode (Enhanced)**: Events are sent via WebSocket connection when available, providing real-time bidirectional communication
- **HTTP Mode (Fallback)**: Events are sent via Ajax/fetch requests with `X-FIR-MODE: 'event'` header when WebSocket is unavailable

Both modes prevent traditional form submission when using `@submit.prevent="$fir.submit()"` and result in DOM updates without page reloads. The same `x-fir-*` attributes and server-side event handlers work identically in both modes.


### Directives

`x-fir-mutation-observer` implements the [MutationObserver API](https://developer.mozilla.org/en-US/docs/Web/API/MutationObserver) as an
alpine directive. It allows you to observe changes to the DOM and react to them.

**Available modifiers:**

- `.child-list` - Monitor for addition/removal of child nodes
- `.attributes` - Monitor for attribute value changes  
- `.subtree` - Extend monitoring to entire subtree
- `.character-data` - Monitor for character data changes
- `.attribute-old-value` - Record previous attribute values
- `.character-data-old-value` - Record previous character data values
- `.attribute-filter:attr1,attr2` - Only monitor specific attributes

e.g.

```html
<!-- Basic usage -->
x-fir-mutation-observer.child-list.subtree="if ($el.children.length === 0) { empty = true } else { empty = false }"

<!-- Monitor specific attributes with old values -->
x-fir-mutation-observer.attributes.attribute-old-value.attribute-filter:class,data-status="handleAttributeChange()"

<!-- Monitor character data changes -->
x-fir-mutation-observer.character-data.character-data-old-value="handleTextChange()"
```

### Magics

See [api.md](api.md)


### Modifiers

Fir supports standard Alpine.js modifiers like `.prevent`, `.stop`, `.once`, etc.

```html
 @fir:create:ok.prevent="$el.reset()"

 @fir:delete:ok.once="$fir.removeEl()"
```
