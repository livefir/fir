# New x-fir-redirect Action Implementation

## Summary

I've successfully created a new magic function and corresponding action to replace the `x-fir-runjs` + `x-fir-js` pattern for navigation with a cleaner, more semantic `x-fir-redirect` action.

## Changes Made

### 1. Alpine.js Plugin - New Magic Function

Added `redirect` magic function in `/alpinejs-plugin/src/magicFunctions.js`:

```javascript
const redirect = (url = '/') => {
    return function (event) {
        // Allow the URL to be passed either as parameter or via event detail
        const targetUrl = event?.detail?.url || url
        
        if (typeof targetUrl !== 'string') {
            console.error('$fir.redirect() requires a valid URL string')
            return
        }
        
        // Redirect to the specified URL
        window.location.href = targetUrl
    }
}
```

### 2. Go Action Handler

Added `RedirectActionHandler` in `/actions.go`:

```go
// RedirectActionHandler handles x-fir-redirect
type RedirectActionHandler struct{}

func (h *RedirectActionHandler) Name() string    { return "redirect" }
func (h *RedirectActionHandler) Precedence() int { return 90 } // Higher precedence than js actions
func (h *RedirectActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
    // Extract URL from first parameter, default to '/' if not provided
    var url = "'/'"
    if len(info.Params) > 0 && strings.TrimSpace(info.Params[0]) != "" {
        // Ensure the URL is properly quoted for JavaScript
        paramUrl := strings.TrimSpace(info.Params[0])
        if !strings.HasPrefix(paramUrl, "'") && !strings.HasPrefix(paramUrl, "\"") {
            url = fmt.Sprintf("'%s'", paramUrl)
        } else {
            url = paramUrl
        }
    }

    // Create the redirect function call with the URL
    jsAction := fmt.Sprintf("$fir.redirect(%s)", url)
    
    // Use TranslateEventExpression to translate the events, forcing nohtml modifier
    return TranslateEventExpression(info.Value, jsAction, "", "nohtml")
}
```

### 3. Registration

Registered the new action handler in the `init()` function in `/actions.go`.

## Usage Examples

### Before (old pattern):
```html
<form x-fir-runjs:redirectHome="delete:ok"
      x-fir-js:redirectHome="window.location.href = '/'">
    <!-- form content -->
</form>
```

### After (new pattern):
```html
<!-- Redirect to root (default) -->
<form x-fir-redirect="delete:ok">
    <!-- form content -->
</form>

<!-- Redirect to specific URL -->
<form x-fir-redirect:"/home"="delete:ok">
    <!-- form content -->
</form>

<!-- Redirect with multiple events -->
<form x-fir-redirect="delete:ok, cancel:done">
    <!-- form content -->
</form>
```

## Transformation Examples

The new action transforms as follows:

1. `<button x-fir-redirect="delete:ok">Delete</button>`
   → `<button @fir:delete:ok.nohtml="$fir.redirect('/')">Delete</button>`

2. `<button x-fir-redirect:"/home"="delete:ok">Delete</button>`
   → `<button @fir:delete:ok.nohtml="$fir.redirect('/home')">Delete</button>`

3. `<button x-fir-redirect="delete:ok, cancel:done">Action</button>`
   → `<button @fir:[delete:ok,cancel:done].nohtml="$fir.redirect('/')">Action</button>`

## Applied Changes

Updated the fira example in `/examples/fira/routes/projects/show.html` to use the new pattern:

```html
<!-- OLD -->
<form x-fir-runjs:redirectHome="delete:ok"
      x-fir-js:redirectHome="window.location.href = '/'">

<!-- NEW -->
<form x-fir-redirect="delete:ok">
```

## Benefits

1. **Cleaner syntax**: Single attribute instead of two
2. **More semantic**: Clear intent with `x-fir-redirect`
3. **Easier to use**: No need to write JavaScript manually
4. **Consistent**: Follows the same pattern as other x-fir-* actions
5. **Default behavior**: Redirects to root by default, but allows customization
6. **Type safety**: Built-in URL validation in the JavaScript function

The implementation is complete and ready for use!
