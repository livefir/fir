package fir

// Page is a struct that holds the data for a page
type Page struct {
	Data    Data   `json:"data"`
	Code    int    `json:"statusCode"`
	Message string `json:"statusMessage"`
	Error   error  `json:"-"`
}

// PageContext is a struct that holds controller data for the page. Its available as `.fir` in the template.
// It provides helper functions for the template.
type PageContext struct {
	Name    string
	URLPath string
}

func (a *PageContext) ActiveRoute(path, class string) string {
	if a.URLPath == path {
		return class
	}
	return ""
}

func (a *PageContext) NotActiveRoute(path, class string) string {
	if a.URLPath != path {
		return class
	}
	return ""
}
