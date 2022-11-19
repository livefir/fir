package fir

// Page is a struct that holds the data for a page
type Page struct {
	Data    Data   `json:"data"`
	Code    int    `json:"statusCode"`
	Message string `json:"statusMessage"`
	Error   error  `json:"-"`
}

// AppContext is a struct that holds the data for the app context
type AppContext struct {
	Name    string
	URLPath string
}

func (a *AppContext) ActiveRoute(path, class string) string {
	if a.URLPath == path {
		return class
	}
	return ""
}

func (a *AppContext) NotActiveRoute(path, class string) string {
	if a.URLPath != path {
		return class
	}
	return ""
}
