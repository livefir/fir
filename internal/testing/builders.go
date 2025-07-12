package testing

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// RouteBuilder provides a fluent API for building test routes
type RouteBuilder struct {
	id         string
	content    string
	loadFunc   string
	events     map[string]string
	data       map[string]interface{}
	middleware []string
}

// NewRouteBuilder creates a new route builder
func NewRouteBuilder() *RouteBuilder {
	return &RouteBuilder{
		events:     make(map[string]string),
		data:       make(map[string]interface{}),
		middleware: make([]string, 0),
	}
}

// WithID sets the route ID
func (rb *RouteBuilder) WithID(id string) *RouteBuilder {
	rb.id = id
	return rb
}

// WithContent sets the route content/template
func (rb *RouteBuilder) WithContent(content string) *RouteBuilder {
	rb.content = content
	return rb
}

// WithCounterTemplate creates a simple counter template
func (rb *RouteBuilder) WithCounterTemplate(initialValue int) *RouteBuilder {
	rb.content = `
		<div id="counter" x-fir-refresh="inc,dec">Count: {{.count}}</div>
		<button formaction="/?event=inc">+</button>
		<button formaction="/?event=dec">-</button>
	`
	rb.data["count"] = initialValue
	return rb
}

// WithFormTemplate creates a form template
func (rb *RouteBuilder) WithFormTemplate(fields []string) *RouteBuilder {
	content := `<form x-fir-refresh="submit">`
	for _, field := range fields {
		content += fmt.Sprintf(`
			<input name="%s" value="{{.%s}}" placeholder="%s">
		`, field, field, field)
	}
	content += `
		<button formaction="/?event=submit">Submit</button>
		{{if .error}}<div class="error">{{.error}}</div>{{end}}
		{{if .success}}<div class="success">{{.success}}</div>{{end}}
	</form>`

	rb.content = content

	// Initialize form fields
	for _, field := range fields {
		rb.data[field] = ""
	}
	rb.data["error"] = ""
	rb.data["success"] = ""

	return rb
}

// WithListTemplate creates a list template
func (rb *RouteBuilder) WithListTemplate(itemName string) *RouteBuilder {
	rb.content = fmt.Sprintf(`
		<div id="list" x-fir-refresh="add,remove,clear">
			<h3>{{.title}}</h3>
			<ul>
				{{range .items}}
				<li>{{.}}</li>
				{{end}}
			</ul>
			<input name="new_%s" value="{{.new_%s}}" placeholder="Enter %s">
			<button formaction="/?event=add">Add</button>
			<button formaction="/?event=clear">Clear</button>
		</div>
	`, itemName, itemName, itemName)

	rb.data["title"] = fmt.Sprintf("%s List", itemName)
	rb.data["items"] = []string{}
	rb.data[fmt.Sprintf("new_%s", itemName)] = ""

	return rb
}

// WithLoadFunction sets the OnLoad function code
func (rb *RouteBuilder) WithLoadFunction(funcCode string) *RouteBuilder {
	rb.loadFunc = funcCode
	return rb
}

// WithEvent adds an event handler
func (rb *RouteBuilder) WithEvent(eventName, handlerCode string) *RouteBuilder {
	rb.events[eventName] = handlerCode
	return rb
}

// WithIncrementEvent adds a simple increment event
func (rb *RouteBuilder) WithIncrementEvent(fieldName string) *RouteBuilder {
	code := fmt.Sprintf(`
		%s++
		ctx.KV("%s", %s)
		return nil
	`, fieldName, fieldName, fieldName)
	return rb.WithEvent("inc", code)
}

// WithDecrementEvent adds a simple decrement event
func (rb *RouteBuilder) WithDecrementEvent(fieldName string) *RouteBuilder {
	code := fmt.Sprintf(`
		%s--
		ctx.KV("%s", %s)
		return nil
	`, fieldName, fieldName, fieldName)
	return rb.WithEvent("dec", code)
}

// WithFormSubmitEvent adds a form submission event
func (rb *RouteBuilder) WithFormSubmitEvent(fields []string, validation map[string]string) *RouteBuilder {
	code := `
		var formData struct {
	`
	for _, field := range fields {
		code += fmt.Sprintf("	%s string `json:\"%s\" form:\"%s\"`\n",
			cases.Title(language.English).String(field), field, field)
	}
	code += `	}
	
		if err := ctx.Bind(&formData); err != nil {
			ctx.KV("error", "Failed to bind form data")
			return nil
		}
		
		// Validation
	`

	for field, validationRule := range validation {
		code += fmt.Sprintf(`
		if formData.%s == "" {
			ctx.KV("error", "%s is required")
			return nil
		}
		`, cases.Title(language.English).String(field), cases.Title(language.English).String(field))

		if validationRule == "email" {
			code += fmt.Sprintf(`
		if !strings.Contains(formData.%s, "@") {
			ctx.KV("error", "Invalid email format")
			return nil
		}
		`, cases.Title(language.English).String(field))
		}
	}

	code += `
		// Success
		ctx.KV("error", "")
		ctx.KV("success", "Form submitted successfully")
	`

	for _, field := range fields {
		code += fmt.Sprintf(`	ctx.KV("%s", formData.%s)`+"\n",
			field, cases.Title(language.English).String(field))
	}

	code += `	return nil`

	return rb.WithEvent("submit", code)
}

// WithData adds data to the route context
func (rb *RouteBuilder) WithData(key string, value interface{}) *RouteBuilder {
	rb.data[key] = value
	return rb
}

// WithMiddleware adds middleware to the route
func (rb *RouteBuilder) WithMiddleware(middleware string) *RouteBuilder {
	rb.middleware = append(rb.middleware, middleware)
	return rb
}

// Build generates the final route configuration code
func (rb *RouteBuilder) Build() string {
	code := fmt.Sprintf(`RouteOptions{
		ID("%s"),
		Content(`+"`%s`"+`),`, rb.id, rb.content)

	if rb.loadFunc != "" {
		code += fmt.Sprintf(`
		OnLoad(func(ctx RouteContext) error {
			%s
		}),`, rb.loadFunc)
	} else if len(rb.data) > 0 {
		code += `
		OnLoad(func(ctx RouteContext) error {`
		for key, value := range rb.data {
			switch v := value.(type) {
			case string:
				code += fmt.Sprintf(`
			ctx.KV("%s", "%s")`, key, v)
			case int:
				code += fmt.Sprintf(`
			ctx.KV("%s", %d)`, key, v)
			default:
				code += fmt.Sprintf(`
			ctx.KV("%s", %v)`, key, v)
			}
		}
		code += `
			return nil
		}),`
	}

	for eventName, handlerCode := range rb.events {
		code += fmt.Sprintf(`
		OnEvent("%s", func(ctx RouteContext) error {
			%s
		}),`, eventName, handlerCode)
	}

	code += `
	}`

	return code
}

// EventBuilder provides a fluent API for building test events
type EventBuilder struct {
	id        string
	params    map[string]interface{}
	isForm    bool
	sessionID string
	timestamp int64
}

// NewEventBuilder creates a new event builder
func NewEventBuilder() *EventBuilder {
	return &EventBuilder{
		params:    make(map[string]interface{}),
		timestamp: time.Now().UTC().UnixMilli(),
	}
}

// WithID sets the event ID
func (eb *EventBuilder) WithID(id string) *EventBuilder {
	eb.id = id
	return eb
}

// WithParam adds a parameter to the event
func (eb *EventBuilder) WithParam(key string, value interface{}) *EventBuilder {
	eb.params[key] = value
	return eb
}

// WithParams sets multiple parameters
func (eb *EventBuilder) WithParams(params map[string]interface{}) *EventBuilder {
	for k, v := range params {
		eb.params[k] = v
	}
	return eb
}

// AsForm marks the event as a form submission
func (eb *EventBuilder) AsForm() *EventBuilder {
	eb.isForm = true
	return eb
}

// WithSession sets the session ID
func (eb *EventBuilder) WithSession(sessionID string) *EventBuilder {
	eb.sessionID = sessionID
	return eb
}

// WithTimestamp sets a custom timestamp
func (eb *EventBuilder) WithTimestamp(timestamp int64) *EventBuilder {
	eb.timestamp = timestamp
	return eb
}

// Build creates the event map
func (eb *EventBuilder) Build() map[string]interface{} {
	event := map[string]interface{}{
		"id":        eb.id,
		"is_form":   eb.isForm,
		"timestamp": eb.timestamp,
	}

	if eb.sessionID != "" {
		event["session_id"] = eb.sessionID
	}

	if len(eb.params) > 0 {
		event["params"] = eb.params
	}

	return event
}

// TemplateBuilder provides utilities for building test templates
type TemplateBuilder struct {
	content    string
	directives []string
	data       map[string]interface{}
}

// NewTemplateBuilder creates a new template builder
func NewTemplateBuilder() *TemplateBuilder {
	return &TemplateBuilder{
		data: make(map[string]interface{}),
	}
}

// WithDiv adds a div element
func (tb *TemplateBuilder) WithDiv(id, content string) *TemplateBuilder {
	tb.content += fmt.Sprintf(`<div id="%s">%s</div>`, id, content)
	return tb
}

// WithButton adds a button element
func (tb *TemplateBuilder) WithButton(text, event string) *TemplateBuilder {
	tb.content += fmt.Sprintf(`<button formaction="/?event=%s">%s</button>`, event, text)
	return tb
}

// WithInput adds an input element
func (tb *TemplateBuilder) WithInput(name, placeholder string) *TemplateBuilder {
	tb.content += fmt.Sprintf(`<input name="%s" value="{{.%s}}" placeholder="%s">`,
		name, name, placeholder)
	return tb
}

// WithFirDirective adds a Fir directive to an element
func (tb *TemplateBuilder) WithFirDirective(directive string) *TemplateBuilder {
	tb.directives = append(tb.directives, directive)
	return tb
}

// WithRefreshDirective adds x-fir-refresh directive
func (tb *TemplateBuilder) WithRefreshDirective(events ...string) *TemplateBuilder {
	directive := fmt.Sprintf(`x-fir-refresh="%s"`, strings.Join(events, ","))
	return tb.WithFirDirective(directive)
}

// WithConditional adds conditional content
func (tb *TemplateBuilder) WithConditional(condition, content string) *TemplateBuilder {
	tb.content += fmt.Sprintf(`{{if %s}}%s{{end}}`, condition, content)
	return tb
}

// WithLoop adds loop content
func (tb *TemplateBuilder) WithLoop(variable, content string) *TemplateBuilder {
	tb.content += fmt.Sprintf(`{{range %s}}%s{{end}}`, variable, content)
	return tb
}

// WithVariable adds a template variable
func (tb *TemplateBuilder) WithVariable(name string) *TemplateBuilder {
	tb.content += fmt.Sprintf(`{{.%s}}`, name)
	return tb
}

// Build generates the final template
func (tb *TemplateBuilder) Build() string {
	if len(tb.directives) > 0 {
		// Wrap content with directives
		directiveStr := strings.Join(tb.directives, " ")
		return fmt.Sprintf(`<div %s>%s</div>`, directiveStr, tb.content)
	}
	return tb.content
}

// ExpectationBuilder provides utilities for building test expectations
type ExpectationBuilder struct {
	statusCode      *int
	htmlContains    []string
	htmlNotContains []string
	jsonContains    map[string]interface{}
	headers         map[string]string
}

// NewExpectationBuilder creates a new expectation builder
func NewExpectationBuilder() *ExpectationBuilder {
	return &ExpectationBuilder{
		htmlContains:    make([]string, 0),
		htmlNotContains: make([]string, 0),
		jsonContains:    make(map[string]interface{}),
		headers:         make(map[string]string),
	}
}

// ExpectStatus sets the expected HTTP status code
func (eb *ExpectationBuilder) ExpectStatus(code int) *ExpectationBuilder {
	eb.statusCode = &code
	return eb
}

// ExpectHTML adds expected HTML content
func (eb *ExpectationBuilder) ExpectHTML(content string) *ExpectationBuilder {
	eb.htmlContains = append(eb.htmlContains, content)
	return eb
}

// ExpectNotHTML adds HTML content that should not be present
func (eb *ExpectationBuilder) ExpectNotHTML(content string) *ExpectationBuilder {
	eb.htmlNotContains = append(eb.htmlNotContains, content)
	return eb
}

// ExpectJSON adds expected JSON data
func (eb *ExpectationBuilder) ExpectJSON(key string, value interface{}) *ExpectationBuilder {
	eb.jsonContains[key] = value
	return eb
}

// ExpectHeader adds expected header
func (eb *ExpectationBuilder) ExpectHeader(name, value string) *ExpectationBuilder {
	eb.headers[name] = value
	return eb
}

// Validate validates a response against the expectations
func (eb *ExpectationBuilder) Validate(t *testing.T, resp *http.Response, body string) {
	if eb.statusCode != nil && resp.StatusCode != *eb.statusCode {
		t.Errorf("Expected status code %d, got %d", *eb.statusCode, resp.StatusCode)
	}

	for _, expected := range eb.htmlContains {
		if !strings.Contains(body, expected) {
			t.Errorf("Response body does not contain expected HTML: %s", expected)
		}
	}

	for _, unexpected := range eb.htmlNotContains {
		if strings.Contains(body, unexpected) {
			t.Errorf("Response body contains unexpected HTML: %s", unexpected)
		}
	}

	for name, expected := range eb.headers {
		actual := resp.Header.Get(name)
		if actual != expected {
			t.Errorf("Expected header %s: %s, got: %s", name, expected, actual)
		}
	}
}

// TestScenarioBuilder provides utilities for building complex test scenarios
type TestScenarioBuilder struct {
	name     string
	steps    []TestStep
	setup    func()
	teardown func()
	data     map[string]interface{}
}

// TestStep represents a single step in a test scenario
type TestStep struct {
	Name        string
	Action      func() interface{}
	Validation  func(result interface{}) error
	Description string
}

// NewTestScenarioBuilder creates a new test scenario builder
func NewTestScenarioBuilder(name string) *TestScenarioBuilder {
	return &TestScenarioBuilder{
		name:  name,
		steps: make([]TestStep, 0),
		data:  make(map[string]interface{}),
	}
}

// WithSetup adds a setup function
func (tsb *TestScenarioBuilder) WithSetup(setup func()) *TestScenarioBuilder {
	tsb.setup = setup
	return tsb
}

// WithTeardown adds a teardown function
func (tsb *TestScenarioBuilder) WithTeardown(teardown func()) *TestScenarioBuilder {
	tsb.teardown = teardown
	return tsb
}

// AddStep adds a test step
func (tsb *TestScenarioBuilder) AddStep(name, description string, action func() interface{}, validation func(interface{}) error) *TestScenarioBuilder {
	step := TestStep{
		Name:        name,
		Action:      action,
		Validation:  validation,
		Description: description,
	}
	tsb.steps = append(tsb.steps, step)
	return tsb
}

// WithData adds shared data for the scenario
func (tsb *TestScenarioBuilder) WithData(key string, value interface{}) *TestScenarioBuilder {
	tsb.data[key] = value
	return tsb
}

// Execute runs the test scenario
func (tsb *TestScenarioBuilder) Execute(t *testing.T) {
	if tsb.setup != nil {
		tsb.setup()
	}

	defer func() {
		if tsb.teardown != nil {
			tsb.teardown()
		}
	}()

	for i, step := range tsb.steps {
		t.Run(fmt.Sprintf("Step_%d_%s", i+1, step.Name), func(t *testing.T) {
			result := step.Action()
			if step.Validation != nil {
				if err := step.Validation(result); err != nil {
					t.Errorf("Step validation failed: %v", err)
				}
			}
		})
	}
}
