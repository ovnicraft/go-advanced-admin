package adminpanel

import (
	"fmt"
	"github.com/ovnicraft/go-advanced-admin/internal/logging"
	"net/http"
	"reflect"
	"strconv"
)

// Model represents a registered model within an app in the admin panel.
type Model struct {
	Name        string
	DisplayName string
	PTR         interface{}
	App         *App
	Fields      []FieldConfig
	ORM         ORMIntegrator
}

// CreateViewLog creates a log entry when the model's list view is accessed.
func (m *Model) CreateViewLog(ctx interface{}) error {
	return m.App.Panel.Config.CreateLog(ctx, logging.LogStoreLevelListView, fmt.Sprintf("%s | %s", m.App.Name, m.DisplayName), nil, "", "")
}

// GetORM returns the ORM integrator for the model.
func (m *Model) GetORM() ORMIntegrator {
	if m.ORM != nil {
		return m.ORM
	}
	return m.App.GetORM()
}

type AdminModelNameInterface interface {
	AdminName() string
}

type AdminModelDisplayNameInterface interface {
	AdminDisplayName() string
}

type AdminModelGetIDInterface interface {
	AdminGetID() interface{}
}

// GetLink returns the relative URL path to the model.
func (m *Model) GetLink() string {
	return fmt.Sprintf("%s/%s", m.App.GetLink(), m.Name)
}

// GetFullLink returns the full URL path to the model, including the admin prefix.
func (m *Model) GetFullLink() string {
	return m.App.Panel.Config.GetLink(m.GetLink())
}

// GetAddLink returns the relative URL path to add a new instance of the model.
func (m *Model) GetAddLink() string {
	return fmt.Sprintf("%s/add", m.GetLink())
}

// GetFullAddLink returns the full URL path to add a new instance of the model.
func (m *Model) GetFullAddLink() string {
	return m.App.Panel.Config.GetLink(m.GetAddLink())
}

// helper: pagination parameters
func getPagination(m *Model, data interface{}) (uint, uint) {
	pageQuery := m.App.Panel.Web.GetQueryParam(data, "page")
	perPageQuery := m.App.Panel.Web.GetQueryParam(data, "perPage")

	var page, perPage uint
	if p, err := strconv.Atoi(pageQuery); err == nil {
		page = uint(p)
	} else {
		page = 1
	}
	if pp, err := strconv.Atoi(perPageQuery); err == nil {
		perPage = uint(pp)
	} else {
		perPage = m.App.Panel.Config.DefaultInstancesPerPage
	}
	if perPage < 10 {
		perPage = 10
	}
	return page, perPage
}

func getFieldsToFetch(m *Model) []string {
	var fields []string
	for _, fc := range m.Fields {
		if fc.IncludeInListFetch {
			fields = append(fields, fc.Name)
		}
	}
	return fields
}

func getFieldsToSearch(m *Model) []string {
	var fields []string
	for _, fc := range m.Fields {
		if fc.IncludeInSearch {
			fields = append(fields, fc.Name)
		}
	}
	return fields
}

func buildCleanInstances(m *Model, data interface{}, instances []interface{}) ([]Instance, error) {
	clean := make([]Instance, len(instances))
	for i, instance := range instances {
		id, err := m.GetPrimaryKeyValue(instance)
		if err != nil {
			return nil, err
		}
		updateAllowed, err := m.App.Panel.PermissionChecker.HasInstanceUpdatePermission(m.App.Name, m.Name, id, data)
		if err != nil {
			return nil, err
		}
		deleteAllowed, err := m.App.Panel.PermissionChecker.HasInstanceDeletePermission(m.App.Name, m.Name, id, data)
		if err != nil {
			return nil, err
		}
		clean[i] = Instance{
			InstanceID:  id,
			Data:        instance,
			Model:       m,
			Permissions: Permissions{Read: true, Update: updateAllowed, Delete: deleteAllowed},
		}
	}
	return clean, nil
}

// GetViewHandler returns the HTTP handler function for the model's list view.
func (m *Model) GetViewHandler() HandlerFunc {
	return func(data interface{}) (uint, string) {
		page, perPage := getPagination(m, data)

		allowed, err := m.App.Panel.PermissionChecker.HasModelReadPermission(m.App.Name, m.Name, data)
		if err != nil {
			return GetErrorHTML(http.StatusInternalServerError, err)
		}
		if !allowed {
			return GetErrorHTML(http.StatusForbidden, fmt.Errorf("forbidden"))
		}

		apps, err := GetAppsWithReadPermissions(m.App.Panel, data)
		if err != nil {
			return GetErrorHTML(http.StatusInternalServerError, err)
		}

		fieldsToFetch := getFieldsToFetch(m)

		searchQuery := m.App.Panel.Web.GetQueryParam(data, "search")
		var instances interface{}
		if searchQuery == "" {
			instances, err = m.GetORM().FetchInstancesOnlyFields(m.PTR, fieldsToFetch)
		} else {
			fieldsToSearch := getFieldsToSearch(m)
			instances, err = m.GetORM().FetchInstancesOnlyFieldWithSearch(m.PTR, fieldsToFetch, searchQuery, fieldsToSearch)
		}
		if err != nil {
			return GetErrorHTML(http.StatusInternalServerError, err)
		}

		filteredInstances, err := filterInstancesByPermission(instances, m, data)
		if err != nil {
			return GetErrorHTML(http.StatusInternalServerError, err)
		}

		totalCount := uint(len(filteredInstances))
		totalPages := (totalCount + perPage - 1) / perPage

		startIndex := (page - 1) * perPage
		endIndex := startIndex + perPage

		if startIndex > totalCount {
			startIndex = totalCount
		}
		if endIndex > totalCount {
			endIndex = totalCount
		}

		pagedInstances := filteredInstances[startIndex:endIndex]

		cleanInstances, err := buildCleanInstances(m, data, pagedInstances)
		if err != nil {
			return GetErrorHTML(http.StatusInternalServerError, err)
		}

		html, err := m.App.Panel.Config.Renderer.RenderTemplate("model", map[string]interface{}{
			"admin":       m.App.Panel,
			"apps":        apps,
			"model":       m,
			"instances":   cleanInstances,
			"totalCount":  totalCount,
			"totalPages":  totalPages,
			"currentPage": page,
			"perPage":     perPage,
			"navBarItems": m.App.Panel.Config.GetNavBarItems(data),
		})
		if err != nil {
			return GetErrorHTML(http.StatusInternalServerError, err)
		}
		err = m.CreateViewLog(data)
		if err != nil {
			return GetErrorHTML(http.StatusInternalServerError, err)
		}
		return http.StatusOK, html
	}
}

// GetPrimaryKeyValue retrieves the primary key value of an instance.
func (m *Model) GetPrimaryKeyValue(instance interface{}) (interface{}, error) {
	return m.GetORM().GetPrimaryKeyValue(instance)
}

// GetPrimaryKeyType retrieves the primary key type of the model.
func (m *Model) GetPrimaryKeyType() (reflect.Type, error) {
	return m.GetORM().GetPrimaryKeyType(m.PTR)
}

func filterInstancesByPermission(instances interface{}, model *Model, data interface{}) ([]interface{}, error) {
	val := reflect.ValueOf(instances)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, fmt.Errorf("instances must be a slice or array")
	}

	filtered := make([]interface{}, 0, val.Len())

	for i := 0; i < val.Len(); i++ {
		instance := val.Index(i).Interface()
		id, err := model.GetPrimaryKeyValue(instance)
		if err != nil {
			return nil, err
		}
		allowed, err := model.App.Panel.PermissionChecker.HasInstanceReadPermission(model.App.Name, model.Name, id, data)
		if err != nil {
			return nil, err
		}
		if allowed && instance != nil {
			filtered = append(filtered, instance)
		}
	}

	return filtered, nil
}

// HandleSearchAJAX handles AJAX search requests for the model
func (m *Model) HandleSearchAJAX(ctx interface{}) error {
	query := m.App.Panel.Web.GetQueryParam(ctx, "q")

	// Get instances from ORM (simplified - would need actual search implementation)
	instances, err := m.GetORM().GetAll(m.PTR)
	if err != nil {
		response := NewErrorResponse([]string{err.Error()})
		return m.App.Panel.Web.SetJSONResponse(ctx, 400, response)
	}

	// Filter instances based on permissions
	filtered, err := filterInstancesByPermission(instances, m, ctx)
	if err != nil {
		response := NewErrorResponse([]string{err.Error()})
		return m.App.Panel.Web.SetJSONResponse(ctx, 400, response)
	}

	// TODO: Implement actual search filtering based on query
	// For now, just return all filtered instances

	response := NewSuccessResponse(map[string]interface{}{
		"instances": filtered,
		"total":     len(filtered),
		"query":     query,
	}, "")

	return m.App.Panel.Web.SetJSONResponse(ctx, 200, response)
}

// HandleDeleteAJAX handles AJAX delete requests for individual instances
func (m *Model) HandleDeleteAJAX(ctx interface{}) error {
	instanceID := m.App.Panel.Web.GetPathParam(ctx, "id")
	if instanceID == "" {
		instanceID = m.App.Panel.Web.GetQueryParam(ctx, "id")
	}

	if instanceID == "" {
		response := NewErrorResponse([]string{"Instance ID is required"})
		return m.App.Panel.Web.SetJSONResponse(ctx, 400, response)
	}

	// Check delete permission
	allowed, err := m.App.Panel.PermissionChecker.HasInstanceDeletePermission(m.App.Name, m.Name, instanceID, ctx)
	if err != nil {
		response := NewErrorResponse([]string{err.Error()})
		return m.App.Panel.Web.SetJSONResponse(ctx, 400, response)
	}

	if !allowed {
		response := NewErrorResponse([]string{"Permission denied"})
		return m.App.Panel.Web.SetJSONResponse(ctx, 403, response)
	}

	// Delete the instance
	err = m.GetORM().DeleteByID(m.PTR, instanceID)
	if err != nil {
		response := NewErrorResponse([]string{err.Error()})
		return m.App.Panel.Web.SetJSONResponse(ctx, 400, response)
	}

	// Create delete log
	instanceIDStr := fmt.Sprintf("%v", instanceID)
	m.App.Panel.Config.CreateLog(ctx, logging.LogStoreLevelInstanceDelete, instanceIDStr, nil, m.Name, "")

	response := NewSuccessResponse(nil, "Item deleted successfully")
	return m.App.Panel.Web.SetJSONResponse(ctx, 200, response)
}

// HandleBulkDeleteAJAX handles AJAX bulk delete requests
func (m *Model) HandleBulkDeleteAJAX(ctx interface{}) error {
	jsonBody, err := m.App.Panel.Web.GetJSONBody(ctx)
	if err != nil {
		response := NewErrorResponse([]string{"Invalid JSON data"})
		return m.App.Panel.Web.SetJSONResponse(ctx, 400, response)
	}

	idsInterface, ok := jsonBody["ids"]
	if !ok {
		response := NewErrorResponse([]string{"No items selected"})
		return m.App.Panel.Web.SetJSONResponse(ctx, 400, response)
	}

	ids, ok := idsInterface.([]interface{})
	if !ok {
		response := NewErrorResponse([]string{"Invalid IDs format"})
		return m.App.Panel.Web.SetJSONResponse(ctx, 400, response)
	}

	deletedCount := 0
	errors := []string{}

	for _, idInterface := range ids {
		id := fmt.Sprintf("%v", idInterface)

		// Check delete permission for each item
		allowed, err := m.App.Panel.PermissionChecker.HasInstanceDeletePermission(m.App.Name, m.Name, id, ctx)
		if err != nil || !allowed {
			errors = append(errors, fmt.Sprintf("Permission denied for item %s", id))
			continue
		}

		// Delete the instance
		err = m.GetORM().DeleteByID(m.PTR, id)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to delete item %s: %s", id, err.Error()))
			continue
		}

		deletedCount++

		// Create delete log
		m.App.Panel.Config.CreateLog(ctx, logging.LogStoreLevelInstanceDelete, fmt.Sprintf("%v", id), nil, m.Name, "")
	}

	if len(errors) > 0 {
		response := JSONResponse{
			Success: deletedCount > 0,
			Message: fmt.Sprintf("%d items deleted, %d failed", deletedCount, len(errors)),
			Data: map[string]interface{}{
				"deleted": deletedCount,
				"failed":  len(errors),
			},
			Errors: errors,
		}
		statusCode := 200
		if deletedCount == 0 {
			statusCode = 400
		}
		return m.App.Panel.Web.SetJSONResponse(ctx, statusCode, response)
	}

	response := NewSuccessResponse(map[string]interface{}{
		"deleted": deletedCount,
	}, fmt.Sprintf("%d items deleted successfully", deletedCount))

	return m.App.Panel.Web.SetJSONResponse(ctx, 200, response)
}
