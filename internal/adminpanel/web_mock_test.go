package adminpanel

type MockWebIntegrator struct{}

func (m *MockWebIntegrator) HandleRoute(string, string, HandlerFunc)         {}
func (m *MockWebIntegrator) HandleJSONRoute(string, string, JSONHandlerFunc) {}
func (m *MockWebIntegrator) ServeAssets(string, TemplateRenderer)            {}
func (m *MockWebIntegrator) GetQueryParam(ctx interface{}, name string) string {
	if query, ok := ctx.(map[string]string); ok {
		return query[name]
	}
	return ""
}
func (m *MockWebIntegrator) GetPathParam(ctx interface{}, name string) string {
	if path, ok := ctx.(map[string]string); ok {
		return path[name]
	}
	return ""
}
func (m *MockWebIntegrator) GetRequestMethod(ctx interface{}) string {
	if request, ok := ctx.(map[string]string); ok {
		return request["method"]
	}
	return ""
}
func (m *MockWebIntegrator) GetFormData(ctx interface{}) map[string][]string {
	if form, ok := ctx.(map[string][]string); ok {
		return form
	}
	return make(map[string][]string)
}

func (m *MockWebIntegrator) SetJSONResponse(ctx interface{}, statusCode int, data interface{}) error {
	return nil
}
func (m *MockWebIntegrator) GetJSONBody(ctx interface{}) (map[string]interface{}, error) {
	if mp, ok := ctx.(map[string]interface{}); ok {
		return mp, nil
	}
	return map[string]interface{}{}, nil
}
