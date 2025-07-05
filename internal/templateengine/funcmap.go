package templateengine

import (
	"html/template"
)

// FuncMapProvider defines an interface for providing template function maps.
// This allows for flexible injection of template functions without tight coupling
// to specific contexts or implementations.
type FuncMapProvider interface {
	// BuildFuncMap creates a template.FuncMap based on the provided context.
	// The context parameter allows for dynamic function generation based on
	// runtime conditions (route context, request data, etc.).
	BuildFuncMap(ctx FuncMapContext) template.FuncMap

	// GetName returns a human-readable name for this provider.
	// Used for debugging and logging.
	GetName() string
}

// FuncMapContext holds the context data needed to build function maps.
// This includes route context, error data, and other runtime information
// that template functions might need access to.
type FuncMapContext struct {
	// RouteContext contains route-specific data (if available)
	RouteContext interface{}

	// Errors contains validation or processing errors to be displayed in templates
	Errors map[string]interface{}

	// URLPath is the current request path (for route-aware functions)
	URLPath string

	// AppName is the application name
	AppName string

	// DevelopmentMode indicates if the app is running in development mode
	DevelopmentMode bool

	// CustomData allows for additional context-specific data
	CustomData map[string]interface{}
}

// DefaultFuncMapProvider implements FuncMapProvider with the standard Fir template functions.
// This includes the 'fir' function that provides access to route DOM context.
type DefaultFuncMapProvider struct {
	name string
}

// NewDefaultFuncMapProvider creates a new default function map provider.
func NewDefaultFuncMapProvider() *DefaultFuncMapProvider {
	return &DefaultFuncMapProvider{
		name: "DefaultFirFuncMap",
	}
}

// BuildFuncMap implements FuncMapProvider interface.
func (dfmp *DefaultFuncMapProvider) BuildFuncMap(ctx FuncMapContext) template.FuncMap {
	return template.FuncMap{
		"fir": func() *RouteDOMContext {
			return NewRouteDOMContext(ctx)
		},
	}
}

// GetName implements FuncMapProvider interface.
func (dfmp *DefaultFuncMapProvider) GetName() string {
	return dfmp.name
}

// CompositeFuncMapProvider combines multiple function map providers.
// Functions from later providers override functions from earlier providers
// when there are naming conflicts.
type CompositeFuncMapProvider struct {
	providers []FuncMapProvider
	name      string
}

// NewCompositeFuncMapProvider creates a new composite function map provider.
func NewCompositeFuncMapProvider(name string, providers ...FuncMapProvider) *CompositeFuncMapProvider {
	return &CompositeFuncMapProvider{
		providers: providers,
		name:      name,
	}
}

// BuildFuncMap implements FuncMapProvider interface.
// Combines function maps from all providers, with later providers taking precedence.
func (cfmp *CompositeFuncMapProvider) BuildFuncMap(ctx FuncMapContext) template.FuncMap {
	result := make(template.FuncMap)

	// Apply function maps in order, allowing later providers to override earlier ones
	for _, provider := range cfmp.providers {
		funcMap := provider.BuildFuncMap(ctx)
		for name, fn := range funcMap {
			result[name] = fn
		}
	}

	return result
}

// GetName implements FuncMapProvider interface.
func (cfmp *CompositeFuncMapProvider) GetName() string {
	return cfmp.name
}

// AddProvider adds a new function map provider to the composite.
func (cfmp *CompositeFuncMapProvider) AddProvider(provider FuncMapProvider) {
	cfmp.providers = append(cfmp.providers, provider)
}

// FuncMapRegistry manages a collection of named function map providers.
// This allows for runtime registration and retrieval of function map providers.
type FuncMapRegistry struct {
	providers map[string]FuncMapProvider
	default_  FuncMapProvider
}

// NewFuncMapRegistry creates a new function map registry.
func NewFuncMapRegistry() *FuncMapRegistry {
	return &FuncMapRegistry{
		providers: make(map[string]FuncMapProvider),
		default_:  NewDefaultFuncMapProvider(),
	}
}

// Register adds a function map provider to the registry.
func (fmr *FuncMapRegistry) Register(name string, provider FuncMapProvider) {
	fmr.providers[name] = provider
}

// Get retrieves a function map provider by name.
// Returns the default provider if the named provider is not found.
func (fmr *FuncMapRegistry) Get(name string) FuncMapProvider {
	if provider, exists := fmr.providers[name]; exists {
		return provider
	}
	return fmr.default_
}

// GetDefault returns the default function map provider.
func (fmr *FuncMapRegistry) GetDefault() FuncMapProvider {
	return fmr.default_
}

// SetDefault sets the default function map provider.
func (fmr *FuncMapRegistry) SetDefault(provider FuncMapProvider) {
	fmr.default_ = provider
}

// List returns all registered provider names.
func (fmr *FuncMapRegistry) List() []string {
	names := make([]string, 0, len(fmr.providers))
	for name := range fmr.providers {
		names = append(names, name)
	}
	return names
}
