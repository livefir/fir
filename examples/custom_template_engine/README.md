# Custom Template Engine Example

This example demonstrates how to create and use a custom template engine with the Fir framework. The custom template engine extends the base functionality with additional features and processing.

## Features Demonstrated

### 1. **Custom Template Processing**
- Adds a custom header to all templates
- Provides custom template functions
- Implements custom caching strategies

### 2. **Extended Functionality**
- `customPrefix()` - Returns the engine's prefix
- `toUpper(string)` - Converts text to uppercase  
- `customFormat(string)` - Formats text with custom prefix
- `errorStyle(string)` - Adds error styling to messages

### 3. **Custom Caching**
- Implements prefixed cache keys
- Dual-layer caching (custom + base engine)
- Custom cache management

## Running the Example

```bash
cd examples/custom_template_engine
go run main.go
```

Then visit [http://localhost:3001](http://localhost:3001) to see the custom template engine in action.

## Implementation Details

### CustomTemplateEngine Structure

```go
type CustomTemplateEngine struct {
    baseEngine templateengine.TemplateEngine
    cache      map[string]templateengine.Template
    cacheMutex sync.RWMutex
    prefix     string
}
```

### Key Methods

- **LoadTemplate**: Adds custom prefix to template content
- **LoadTemplateWithContext**: Injects custom template functions
- **CacheTemplate**: Uses prefixed cache keys
- **Custom Processing**: Adds headers and styling

## Custom Functions Available in Templates

- `{{customPrefix}}` - Displays the engine prefix
- `{{toUpper "text"}}` - Converts text to uppercase
- `{{customFormat "text"}}` - Formats text with prefix
- `{{errorStyle "message"}}` - Styles error messages

## Use Cases

This pattern is useful for:

1. **Multi-tenant Applications**: Different styling per tenant
2. **A/B Testing**: Different template processing per variant
3. **Feature Flags**: Conditional template processing
4. **Branding**: Organization-specific template modifications
5. **Performance Optimization**: Custom caching strategies
6. **Content Transformation**: Pre/post-processing of template content

## Extension Points

The custom template engine can be extended to:

- Add template validation
- Implement template versioning
- Add template analytics
- Integrate with external template systems
- Implement template compilation optimizations
- Add template security features
