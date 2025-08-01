---
applyTo: '**'
---

# Fir Actions: HTTP vs WebSocket Mode Behavior

This document explains how Fir actions work in two distinct modes and provides guidance for implementation and testing based on the Fir framework architecture.

## Core Architecture Understanding

Fir operates with a dual-mode architecture that provides progressive enhancement:

- **HTTP Mode (Fallback)**: Event submission via Ajax/fetch with DOM updates (no page reloads when using @submit.prevent)
- **WebSocket Mode (Enhanced)**: Event submission via WebSocket with real-time DOM updates

Both modes use identical server-side code and `x-fir-*` attributes, ensuring consistent behavior regardless of client capabilities.

## Two Operational Modes

### HTTP Mode (Fallback)
- **Trigger**: `@submit.prevent="$fir.submit()"` calls when WebSocket is unavailable or fails
- **Transport**: Ajax/fetch requests with `X-FIR-MODE: 'event'` header
- **Data Flow**: `Alpine.js Submit → fetch() → OnEvent handler → writeAndPublishEvents() → JSON response → DOM updates`
- **State Management**: Server state is preserved in closure variables, DOM updated via JSON response events
- **Action Processing**: Server sends DOM events in JSON response, Alpine.js plugin executes corresponding actions
- **Event Publishing**: Uses `writeAndPublishEvents()` which both publishes to WebSocket clients AND writes JSON HTTP response
- **When Used**: When WebSocket is unavailable, connection fails, or as fallback mechanism
- **Key Behavior**: No page reloads occur when using @submit.prevent - DOM is updated via JavaScript like WebSocket mode

### WebSocket Mode (Enhanced)
- **Trigger**: `@submit.prevent="$fir.submit()"` calls when WebSocket connection is active
- **Transport**: WebSocket messages
- **Data Flow**: `Alpine.js Submit → WebSocket → OnEvent handler → publishEvents() → DOM actions → Client updates`
- **State Management**: Same server-side state management, client receives targeted DOM updates via Alpine.js actions
- **Action Processing**: Server sends events to Alpine.js plugin via WebSocket, which executes pre-processed DOM actions
- **Event Publishing**: Uses `publishEvents()` which only publishes to WebSocket subscribers
- **When Used**: When WebSocket connection is active (`window.$fir && window.$fir.ws && window.$fir.ws.readyState === 1`) and Alpine.js is loaded
- **Key Behavior**: Real-time DOM updates without any HTTP requests

## Action Processing Architecture

### Template Processing Pipeline
1. **Raw HTML Template**: Contains `x-fir-*` attributes
2. **Attribute Processing**: `processRenderAttributes()` parses `x-fir-*` attributes
3. **Action Registry**: Each action type has a dedicated `ActionHandler` (11 total handlers)
4. **Code Generation**: Handlers generate Alpine.js `@fir:event` handlers
5. **Template Compilation**: Final template has `@fir:event` handlers instead of `x-fir-*` attributes

### Action Registry (11 Action Types)
```go
var registry = map[string]ActionHandler{
    "refresh":         &RefreshActionHandler{},
    "reset":           &ResetActionHandler{},
    "remove":          &RemoveActionHandler{},
    "remove-parent":   &RemoveParentActionHandler{},
    "append":          &AppendActionHandler{},
    "prepend":         &PrependActionHandler{},
    "toggle-disabled": &ToggleDisabledActionHandler{},
    "toggleClass":     &ToggleClassActionHandler{},
    "dispatch":        &DispatchActionHandler{},
    "runjs":           &TriggerActionHandler{},
    "js":              &ActionPrefixHandler{},
    "redirect":        &RedirectActionHandler{},
}
```

## Action-Specific Behavior Differences

### x-fir-refresh
- **HTTP Mode**: Element content updated via JSON response events
- **WebSocket Mode**: Element content updated via WebSocket server events

### x-fir-append/prepend
- **HTTP Mode**: New DOM elements added via JSON response processing
- **WebSocket Mode**: New DOM elements added via WebSocket message processing

### x-fir-remove/remove-parent
- **HTTP Mode**: Element removed from DOM via JSON response events
- **WebSocket Mode**: Element removed from DOM via WebSocket events

### x-fir-reset
- **HTTP Mode**: Form reset via JSON response action processing
- **WebSocket Mode**: Form reset via WebSocket action processing

### x-fir-toggle-disabled
- **HTTP Mode**: Elements disabled/enabled via JSON response events
- **WebSocket Mode**: Elements disabled/enabled via WebSocket events

### x-fir-toggleClass
- **HTTP Mode**: CSS classes toggled via JSON response action processing
- **WebSocket Mode**: CSS classes toggled via WebSocket action processing

### x-fir-dispatch
- **HTTP Mode**: Custom events dispatched via JSON response processing
- **WebSocket Mode**: Custom events dispatched via WebSocket processing

### x-fir-js/runjs
- **HTTP Mode**: JavaScript executed via JSON response action processing
- **WebSocket Mode**: JavaScript executed via WebSocket action processing

### x-fir-redirect
- **HTTP Mode**: Client-side navigation via JSON response
- **WebSocket Mode**: Client-side navigation via WebSocket message

## Template Structure Requirements

### Dual Mode Support
Forms should support both modes for proper fallback:

```html
<!-- WebSocket mode with HTTP fallback -->
<form method="post" @submit.prevent="$fir.submit()">
    <button class="button" formaction="/?event=add-item" type="submit" 
            x-fir-append:item="add-item">Add Item</button>
</form>
```

**Key Pattern**: The `@submit.prevent="$fir.submit()"` prevents traditional form submission and uses JavaScript (either WebSocket or Ajax/fetch) for event handling. Both modes result in DOM updates without page reloads.

### Block Templates for Dynamic Content
Use Go template blocks for content that needs dynamic updates:

```html
<ul id="items-list" x-fir-append:item="add-item">
    {{ range .items }}
        {{ block "item" . }}
            <li class="item">{{ . }}</li>
        {{ end }}
    {{ end }}
</ul>
```

**Note**: The @fir event handlers (like `@fir:create:ok::item="$fir.appendEl()"`) are internal implementation details created by Fir's template processing. They should only be referenced for debugging purposes, not for development.

## Testing Strategy

### Comprehensive E2E Testing
Tests must cover both modes to ensure complete functionality:

1. **HTTP Mode Tests**: Test Ajax/fetch based event handling
2. **WebSocket Mode Tests**: Test `$fir.submit()` and real-time updates
3. **Fallback Tests**: Test Ajax fallback when WebSocket is unavailable
4. **Behavior Comparison**: Verify both modes achieve same end result
5. **State Persistence Verification**: In both modes, verify server state is updated and DOM reflects changes

### Test Implementation Patterns

```go
func TestActionsE2E(t *testing.T) {
    t.Run("HTTPMode", func(t *testing.T) {
        // Test Ajax/fetch based event handling
        testAllActionsHTTPMode(t, ctx, baseURL)
    })
    
    t.Run("WebSocketMode", func(t *testing.T) {
        // Wait for WebSocket connection
        chromedp.Poll(`window.$fir && window.$fir.ws && window.$fir.ws.readyState === 1`, nil)
        testAllActionsWebSocketMode(t, ctx, baseURL)
    })
    
    t.Run("HTTPFallback", func(t *testing.T) {
        // Disable WebSocket and test Ajax fallback
        testHTTPFallbackWhenWebSocketUnavailable(t, ctx, baseURL)
    })
}
```

## Server-Side Implementation

### OnLoad Data Method
Use `ctx.Data()` for initial page data instead of multiple `ctx.KV()` calls:

```go
fir.OnLoad(func(ctx fir.RouteContext) error {
    return ctx.Data(map[string]interface{}{
        "counter": counter,
        "items": items,
        "message": message,
    })
})
```

### Event Handlers
Return appropriate data for both modes:

```go
fir.OnEvent("add-item", func(ctx fir.RouteContext) error {
    // Update server state
    items = append(items, newItem)
    
    // Return data for both HTTP and WebSocket updates
    return ctx.KV("items", items)
})
```

## Development Guidelines

### Progressive Enhancement
1. Start with HTTP mode functionality (forms with `formaction`)
2. Add WebSocket mode enhancements (`@submit.prevent="$fir.submit()"`)
3. Ensure fallback works when JavaScript is disabled
4. Test both modes thoroughly

### Action Attribute Patterns
- Use consistent event names across HTTP and WebSocket modes
- Include template blocks for dynamic content
- Provide proper CSS selectors for DOM targeting
- Handle error states in both modes

### Common Pitfalls
- Don't reset persistent state on every page load (use conditional initialization)
- Ensure WebSocket mode doesn't break when connection is lost
- Test form submissions work without JavaScript
- Verify action selectors match actual DOM structure

## Debugging Tips

### WebSocket Connection Issues
Check WebSocket status: `window.$fir && window.$fir.ws && window.$fir.ws.readyState === 1`

### Action Processing
Monitor console for Fir events and action execution logs

### Template Rendering
Verify server-side template context includes all necessary data

### DOM Updates
Use browser dev tools to confirm DOM changes occur as expected in both modes

## Event Publishing Flow

### HTTP Mode Publishing
```go
func writeAndPublishEvents(ctx RouteContext) eventPublisher {
    return func(pubsubEvent pubsub.Event) error {
        // Publish to WebSocket subscribers
        ctx.route.pubsub.Publish(ctx.request.Context(), channel, pubsubEvent)
        // Write JSON response for HTTP client
        return writeEventHTTP(ctx, pubsubEvent)
    }
}
```

### WebSocket Mode Publishing
```go
func publishEvents(ctx context.Context, eventCtx RouteContext, channel string) eventPublisher {
    return func(pubsubEvent pubsub.Event) error {
        return eventCtx.route.pubsub.Publish(ctx, channel, pubsubEvent)
    }
}
```

## Key Architecture Insights

1. **Same Server Logic**: Both modes use identical OnLoad/OnEvent handlers
2. **Different Client Handling**: HTTP uses Ajax/fetch, WebSocket uses real-time connection
3. **Progressive Enhancement**: WebSocket mode enhances HTTP baseline
4. **State Consistency**: Server state management is identical in both modes
5. **Event Broadcasting**: All events are broadcast to WebSocket clients regardless of origin mode
6. **Template Processing**: x-fir-* attributes are processed at compile time, not runtime
7. **Action Registry**: All 11 action types work in both modes with appropriate adaptations
8. **No Page Reloads**: Both modes prevent page reloads when using @submit.prevent="$fir.submit()"
