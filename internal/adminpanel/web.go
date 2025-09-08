package adminpanel

// HandlerFunc represents a handler function used in the admin panel routes.
type HandlerFunc = func(interface{}) (uint, string)

// JSONHandlerFunc represents a handler function for JSON responses.
type JSONHandlerFunc = func(interface{}) error

// WebIntegrator defines the interface for integrating web frameworks with the admin panel.
type WebIntegrator interface {
	// HandleRoute registers a route with the given method, path, and handler function.
	HandleRoute(method, path string, handler HandlerFunc)

	// HandleJSONRoute registers a route that returns JSON responses.
	HandleJSONRoute(method, path string, handler JSONHandlerFunc)

	// ServeAssets serves static assets under the specified prefix using the provided renderer.
	ServeAssets(prefix string, renderer TemplateRenderer)

	// GetQueryParam retrieves the value of a query parameter from the context.
	GetQueryParam(ctx interface{}, name string) string

	// GetPathParam retrieves the value of a path parameter from the context.
	GetPathParam(ctx interface{}, name string) string

	// GetRequestMethod retrieves the HTTP method of the request from the context.
	GetRequestMethod(ctx interface{}) string

	// GetFormData retrieves form data from the context.
	GetFormData(ctx interface{}) map[string][]string

	// SetJSONResponse sets a JSON response with the given status code and data.
	SetJSONResponse(ctx interface{}, statusCode int, data interface{}) error

	// GetJSONBody retrieves JSON body data from the request context.
	GetJSONBody(ctx interface{}) (map[string]interface{}, error)
}

// JSONResponse represents a standard JSON response structure.
type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
}

// NewSuccessResponse creates a success JSON response.
func NewSuccessResponse(data interface{}, message string) JSONResponse {
	return JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates an error JSON response.
func NewErrorResponse(errors []string) JSONResponse {
	return JSONResponse{
		Success: false,
		Errors:  errors,
	}
}
