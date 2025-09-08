package adminpanel

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/ovnicraft/go-advanced-admin/internal/form"
	"github.com/ovnicraft/go-advanced-admin/internal/form/fields"
	"github.com/ovnicraft/go-advanced-admin/internal/logging"
	"github.com/ovnicraft/go-advanced-admin/internal/utils"
	"net/http"
	"reflect"
	"strings"
)

// App represents an application within the admin panel, grouping related models together.
type App struct {
	Name        string
	DisplayName string
	Models      map[string]*Model
	ModelsSlice []*Model
	Panel       *AdminPanel
	ORM         ORMIntegrator
}

// CreateViewLog creates a log entry when the app is viewed.
func (a *App) CreateViewLog(ctx interface{}) error {
	return a.Panel.Config.CreateLog(ctx, logging.LogStoreLevelPanelView, a.Name, nil, "", "")
}

// GetORM returns the ORM integrator for the app.
func (a *App) GetORM() ORMIntegrator {
	if a.ORM != nil {
		return a.ORM
	}
	return a.Panel.GetORM()
}

// RegisterModel registers a model with the app, making it available in the admin interface.
func (a *App) RegisterModel(model interface{}, orm ORMIntegrator) (*Model, error) {
	modelType := reflect.TypeOf(model)

	if modelType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("admin model '%s' must be a pointer to a struct", modelType.Name())
	}

	modelType = modelType.Elem()
	if modelType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("admin model '%s' must be a pointer to a struct", modelType.Name())
	}

	var name string
	namer, ok := model.(AdminModelNameInterface)
	if ok {
		name = namer.AdminName()
	} else {
		name = modelType.Name()
	}

	if !utils.IsURLSafe(name) {
		return nil, fmt.Errorf("admin model '%s' name is not URL safe", name)
	}

	var displayName string
	displayNamer, ok := model.(AdminModelDisplayNameInterface)
	if ok {
		displayName = displayNamer.AdminDisplayName()
	} else {
		displayName = utils.HumanizeName(name)
	}

	if _, exists := a.Models[name]; exists {
		return nil, fmt.Errorf("admin model '%s' already exists in app '%s'. Models cannot be registered more than once", name, a.Name)
	}

	var fieldConfigs []FieldConfig
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		fieldName := field.Name
		fieldType := field.Type

		isPointer := false
		underlyingType := fieldType
		if fieldType.Kind() == reflect.Ptr {
			isPointer = true
			underlyingType = fieldType.Elem()
		}

		tag := field.Tag.Get("admin")
		opts, err := parseInclusionTags(tag, fieldName)
		if err != nil {
			return nil, err
		}

		var formField form.Field
		if opts.includeInAddForm || opts.includeInEditForm {
			formField, err = buildFormField(underlyingType, fieldType, tag)
			if err != nil {
				return nil, err
			}
			if formField != nil {
				if err := applyInitialValueTag(formField, tag, underlyingType); err != nil {
					return nil, err
				}
			}
		}

		var formAddField, formEditField form.Field
		if opts.includeInAddForm {
			formAddField = formField
		}
		if opts.includeInEditForm {
			formEditField = formField
		}

		if gen, ok := model.(AdminFormFieldInterface); ok {
			if ff := gen.AdminFormField(fieldName, false); ff != nil {
				formAddField = ff
			}
			if ff := gen.AdminFormField(fieldName, true); ff != nil {
				formEditField = ff
			}
		}

		fieldConfigs = append(fieldConfigs, FieldConfig{
			Name:                  fieldName,
			DisplayName:           opts.fieldDisplayName,
			FieldType:             underlyingType,
			IsPointer:             isPointer,
			IncludeInListDisplay:  opts.includeInList,
			IncludeInListFetch:    opts.includeInFetch,
			IncludeInSearch:       opts.includeInSearch,
			IncludeInInstanceView: opts.includeInInstanceView,
			AddFormField:          formAddField,
			EditFormField:         formEditField,
		})
	}

	modelInstance := &Model{
		Name:        name,
		DisplayName: displayName,
		PTR:         model,
		App:         a,
		Fields:      fieldConfigs,
		ORM:         orm,
	}
	a.Panel.Web.HandleRoute("GET", a.Panel.Config.GetPrefix()+modelInstance.GetLink(), modelInstance.GetViewHandler())
	a.Panel.Web.HandleRoute("GET", a.Panel.Config.GetPrefix()+modelInstance.GetLink()+"/:id/view", modelInstance.GetInstanceViewHandler())
	a.Panel.Web.HandleRoute("DELETE", a.Panel.Config.GetPrefix()+modelInstance.GetLink()+"/:id/view", modelInstance.GetInstanceDeleteHandler())
	a.Panel.Web.HandleRoute("GET", a.Panel.Config.GetPrefix()+modelInstance.GetLink()+"/add", modelInstance.GetAddHandler())
	a.Panel.Web.HandleRoute("POST", a.Panel.Config.GetPrefix()+modelInstance.GetLink()+"/add", modelInstance.GetAddHandler())
	a.Panel.Web.HandleRoute("GET", a.Panel.Config.GetPrefix()+modelInstance.GetLink()+"/:id/edit", modelInstance.GetEditHandler())
	a.Panel.Web.HandleRoute("POST", a.Panel.Config.GetPrefix()+modelInstance.GetLink()+"/:id/edit", modelInstance.GetEditHandler())
	a.ModelsSlice = append(a.ModelsSlice, modelInstance)
	a.Models[name] = modelInstance
	return modelInstance, nil
}

// ---- Helpers to reduce RegisterModel complexity ----

type tagOptions struct {
	includeInList         bool
	includeInFetch        bool
	includeInSearch       bool
	includeInInstanceView bool
	includeInAddForm      bool
	includeInEditForm     bool
	fieldDisplayName      string
}

func forEachTag(tag string, fn func(key, value string)) {
	if tag == "" {
		return
	}
	parts := strings.Split(tag, ";")
	for _, t := range parts {
		kv := strings.SplitN(t, ":", 2)
		if len(kv) == 2 {
			fn(kv[0], kv[1])
		} else if len(kv) == 1 {
			fn(kv[0], "")
		}
	}
}

func parseInclusionTags(tag, fieldName string) (tagOptions, error) {
	opts := tagOptions{
		includeInList:         true,
		includeInFetch:        true,
		includeInSearch:       true,
		includeInInstanceView: true,
		includeInAddForm:      true,
		includeInEditForm:     true,
		fieldDisplayName:      utils.HumanizeName(fieldName),
	}

	listFetchTagPresent := false
	var retErr error
	forEachTag(tag, func(key, value string) {
		switch key {
		case "listDisplay":
			if value == "include" {
				opts.includeInList = true
			} else if value == "exclude" {
				opts.includeInList = false
			} else if value != "" { // invalid explicit value
				retErr = fmt.Errorf("invalid value for 'listDisplay' tag: %s", value)
			}
		case "listFetch":
			listFetchTagPresent = true
			if value == "include" {
				opts.includeInFetch = true
			} else if value == "exclude" {
				opts.includeInFetch = false
			} else {
				retErr = fmt.Errorf("invalid value for 'listFetch' tag: %s", value)
			}
		case "search":
			if value == "include" {
				opts.includeInSearch = true
			} else if value == "exclude" {
				opts.includeInSearch = false
			} else {
				retErr = fmt.Errorf("invalid value for 'search' tag: %s", value)
			}
		case "view":
			if value == "include" {
				opts.includeInInstanceView = true
			} else if value == "exclude" {
				opts.includeInInstanceView = false
			} else {
				retErr = fmt.Errorf("invalid value for 'view' tag: %s", value)
			}
		case "addForm":
			if value == "include" {
				opts.includeInAddForm = true
			} else if value == "exclude" {
				opts.includeInAddForm = false
			} else {
				retErr = fmt.Errorf("invalid value for 'addForm' tag: %s", value)
			}
		case "editForm":
			if value == "include" {
				opts.includeInEditForm = true
			} else if value == "exclude" {
				opts.includeInEditForm = false
			} else {
				retErr = fmt.Errorf("invalid value for 'editForm' tag: %s", value)
			}
		case "displayName":
			opts.fieldDisplayName = value
		}
	})

	if !listFetchTagPresent {
		if fieldName == "ID" {
			opts.includeInFetch = true
		} else {
			opts.includeInFetch = opts.includeInList
		}
	}
	if retErr != nil {
		return opts, retErr
	}
	return opts, nil
}

func buildFormField(underlyingType reflect.Type, fieldType reflect.Type, tag string) (form.Field, error) {
	switch underlyingType.Kind() {
	case reflect.String:
		return configureTextField(tag), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return configureIntegerField(tag), nil
	case reflect.Float32, reflect.Float64:
		return configureFloatField(tag), nil
	case reflect.Bool:
		return configureBooleanField(tag), nil
	default:
		return configureUUIDField(tag, fieldType == reflect.TypeOf(uuid.UUID{})), nil
	}
}

func configureTextField(tag string) *fields.TextField {
	tf := &fields.TextField{}
	forEachTag(tag, func(key, value string) {
		switch key {
		case "placeholder":
			tf.Placeholder = &value
		case "required":
			tf.Required = true
		case "regex":
			tf.Regex = &value
		case "maxLength":
			if v, err := utils.ConvertStringToType(value, reflect.TypeOf(uint(0))); err == nil {
				if vv, ok := v.(uint); ok {
					tf.MaxLength = &vv
				}
			}
		case "minLength":
			if v, err := utils.ConvertStringToType(value, reflect.TypeOf(uint(0))); err == nil {
				if vv, ok := v.(uint); ok {
					tf.MinLength = &vv
				}
			}
		}
	})
	return tf
}

func configureIntegerField(tag string) *fields.IntegerField {
	tf := &fields.IntegerField{}
	forEachTag(tag, func(key, value string) {
		switch key {
		case "required":
			tf.Required = true
		case "max":
			if v, err := utils.ConvertStringToType(value, reflect.TypeOf(0)); err == nil {
				if vv, ok := v.(int); ok {
					tf.MaxValue = &vv
				}
			}
		case "min":
			if v, err := utils.ConvertStringToType(value, reflect.TypeOf(0)); err == nil {
				if vv, ok := v.(int); ok {
					tf.MinValue = &vv
				}
			}
		}
	})
	return tf
}

func configureFloatField(tag string) *fields.FloatField {
	tf := &fields.FloatField{}
	forEachTag(tag, func(key, value string) {
		switch key {
		case "required":
			tf.Required = true
		case "max":
			if v, err := utils.ConvertStringToType(value, reflect.TypeOf(float64(0))); err == nil {
				if vv, ok := v.(float64); ok {
					tf.MaxValue = &vv
				}
			}
		case "min":
			if v, err := utils.ConvertStringToType(value, reflect.TypeOf(float64(0))); err == nil {
				if vv, ok := v.(float64); ok {
					tf.MinValue = &vv
				}
			}
		}
	})
	return tf
}

func configureBooleanField(tag string) *fields.BooleanField {
	tf := &fields.BooleanField{}
	forEachTag(tag, func(key, _ string) {
		if key == "required" {
			tf.Required = true
		}
	})
	return tf
}

func configureUUIDField(tag string, isUUID bool) *fields.UUIDField {
	tf := &fields.UUIDField{}
	if isUUID {
		forEachTag(tag, func(key, _ string) {
			if key == "required" {
				tf.Required = true
			}
		})
	}
	return tf
}

func applyInitialValueTag(f form.Field, tag string, typ reflect.Type) error {
	var convErr error
	forEachTag(tag, func(key, value string) {
		if key == "initial" {
			v, err := utils.ConvertStringToType(value, typ)
			if err != nil {
				convErr = fmt.Errorf("error converting value '%s' to type '%s': %w", value, typ.Name(), err)
				return
			}
			f.RegisterInitialValue(v)
		}
	})
	return convErr
}

// GetHandler returns the HTTP handler function for the app's main page.
func (a *App) GetHandler() HandlerFunc {
	return func(data interface{}) (uint, string) {
		allowed, err := a.Panel.PermissionChecker.HasAppReadPermission(a.Name, data)
		if err != nil {
			return GetErrorHTML(http.StatusInternalServerError, err)
		}
		if !allowed {
			return GetErrorHTML(http.StatusForbidden, fmt.Errorf("forbidden"))
		}

		models, err := GetModelsWithReadPermissions(a, data)
		if err != nil {
			return GetErrorHTML(http.StatusInternalServerError, err)
		}

		html, err := a.Panel.Config.Renderer.RenderTemplate("app", map[string]interface{}{"admin": a.Panel, "app": a, "models": models, "navBarItems": a.Panel.Config.GetNavBarItems(data)})
		if err != nil {
			return GetErrorHTML(http.StatusInternalServerError, err)
		}
		err = a.CreateViewLog(data)
		if err != nil {
			return GetErrorHTML(http.StatusInternalServerError, err)
		}
		return http.StatusOK, html
	}
}

// GetLink returns the relative URL path to the app.
func (a *App) GetLink() string {
	return fmt.Sprintf("/a/%s", a.Name)
}

// GetFullLink returns the full URL path to the app, including the admin prefix.
func (a *App) GetFullLink() string {
	return a.Panel.Config.GetLink(a.GetLink())
}
