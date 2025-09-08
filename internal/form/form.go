package form

import (
	"fmt"
	"html/template"
	"strings"
)

type ValidationFunc func(map[string]interface{}) (frontend []error, backend error)

type Form interface {
	AddField(name string, field Field) error
	GetFields() []Field
	RegisterValidationFunctions(validationFuncs ...ValidationFunc)
	GetValidationFunctions() []ValidationFunc
	RegisterInitialValues(values map[string]interface{}) error
	Save(values map[string]HTMLType) (interface{}, error)
}

func ValuesAreValid(form Form, values map[string]interface{}) ([]error, map[string][]error, error) {
	formErrs := make([]error, 0)
	fieldsErrs := make(map[string][]error)

	fields := form.GetFields()
	for _, field := range fields {
		fieldName := field.GetName()
		fieldValue, exists := values[fieldName]
		if !exists {
			fieldValue = nil
		}
		fieldErrs, err := FieldValueIsValid(field, fieldValue)
		if err != nil {
			return formErrs, fieldsErrs, err
		}
		fieldsErrs[fieldName] = fieldErrs
	}

	validationFuncs := form.GetValidationFunctions()
	for _, validationFunc := range validationFuncs {
		frontend, err := validationFunc(values)
		if err != nil {
			return formErrs, fieldsErrs, err
		}
		if frontend != nil {
			formErrs = append(formErrs, frontend...)
		}
	}

	return formErrs, fieldsErrs, nil
}

func GetCleanData(form Form, values map[string]HTMLType) (map[string]interface{}, error) {
	cleanValues := make(map[string]interface{})

	for _, field := range form.GetFields() {
		fieldName := field.GetName()
		fieldValue, exists := values[fieldName]
		if !exists {
			fieldValue = ""
		}
		cleanValue, err := field.HTMLTypeToGoType(fieldValue)
		if err != nil {
			return nil, err
		}
		cleanValues[fieldName] = cleanValue
	}

	return cleanValues, nil
}

func renderErrors(errors []error) string {
	if len(errors) == 0 {
		return ""
	}
	var errStrings []string
	for _, err := range errors {
		errStrings = append(errStrings, template.HTMLEscapeString(err.Error()))
	}
	return fmt.Sprintf(`<ul class="errorlist"><li>%s</li></ul>`, strings.Join(errStrings, "</li><li>"))
}

func RenderFormAsP(form Form, formErrs []error, fieldsErrs map[string][]error) (string, error) {
	var htmlStrings []string
	for _, field := range form.GetFields() {
		fieldHTML, err := field.HTML()
		if err != nil {
			return "", err
		}
		label := template.HTMLEscapeString(field.GetLabel())
		fieldErrs, exists := fieldsErrs[field.GetName()]
		fieldErrors := ""
		if exists && len(fieldErrs) > 0 {
			fieldErrors = renderErrors(fieldErrs)
		}
		htmlStrings = append(htmlStrings, fmt.Sprintf(`<p><label for="%s">%s:</label> %s%s</p>`, field.GetName(), label, fieldHTML, fieldErrors))
	}
	if len(formErrs) > 0 {
		htmlStrings = append(htmlStrings, renderErrors(formErrs))
	}
	return strings.Join(htmlStrings, "\n"), nil
}

func RenderFormAsUL(form Form, formErrs []error, fieldsErrs map[string][]error) (string, error) {
	var htmlStrings []string
	for _, field := range form.GetFields() {
		fieldHTML, err := field.HTML()
		if err != nil {
			return "", err
		}
		label := template.HTMLEscapeString(field.GetLabel())
		fieldErrs, exists := fieldsErrs[field.GetName()]
		fieldErrors := ""
		if exists && len(fieldErrs) > 0 {
			fieldErrors = renderErrors(fieldErrs)
		}
		htmlStrings = append(htmlStrings, fmt.Sprintf(`<li><label for="%s">%s:</label> %s%s</li>`, field.GetName(), label, fieldHTML, fieldErrors))
	}
	if len(formErrs) > 0 {
		htmlStrings = append(htmlStrings, renderErrors(formErrs))
	}
	return fmt.Sprintf("<ul>\n%s\n</ul>", strings.Join(htmlStrings, "\n")), nil
}

func RenderFormAsTable(form Form, formErrs []error, fieldsErrs map[string][]error) (string, error) {
	var htmlStrings []string
	for _, field := range form.GetFields() {
		fieldHTML, err := field.HTML()
		if err != nil {
			return "", err
		}
		label := template.HTMLEscapeString(field.GetLabel())
		fieldErrs, exists := fieldsErrs[field.GetName()]
		fieldErrors := ""
		if exists && len(fieldErrs) > 0 {
			fieldErrors = renderErrors(fieldErrs)
		}
		htmlStrings = append(htmlStrings, fmt.Sprintf(`<tr><th><label for="%s">%s</label></th><td>%s%s</td></tr>`, field.GetName(), label, fieldHTML, fieldErrors))
	}
	if len(formErrs) > 0 {
		htmlStrings = append(htmlStrings, fmt.Sprintf("<tr><td colspan=\"2\">%s</td></tr>", renderErrors(formErrs)))
	}
	return fmt.Sprintf("<table>\n%s\n</table>", strings.Join(htmlStrings, "\n")), nil
}

func RenderFormAsTabler(form Form, formErrs []error, fieldsErrs map[string][]error) (string, error) {
    var htmlStrings []string
    for _, field := range form.GetFields() {
        fieldHTML, err := field.HTML()
        if err != nil {
            return "", err
        }
        label := template.HTMLEscapeString(field.GetLabel())
        fieldErrs, exists := fieldsErrs[field.GetName()]

		// Add required indicator and error styling
		labelClass := "form-label"

		// Check if field has validation functions (indicating it's required)
		validationFuncs := field.GetValidationFunctions()
		if len(validationFuncs) > 0 {
			labelClass = "form-label required-label"
		}

        // Special handling for boolean checkbox fields
        if strings.Contains(fieldHTML, `type="checkbox"`) {
            // Ensure input has id and proper class
            name := field.GetName()
            // Inject id attribute if missing
            if !strings.Contains(fieldHTML, " id=") {
                fieldHTML = strings.Replace(fieldHTML, "<input", `<input id="`+name+`"`, 1)
            }
            // Ensure checkbox class
            if strings.Contains(fieldHTML, `class="`) {
                fieldHTML = strings.ReplaceAll(fieldHTML, `class="form-control"`, `class="form-check-input"`)
            } else {
                fieldHTML = strings.Replace(fieldHTML, `id="`+name+`"`, `id="`+name+`" class="form-check-input"`, 1)
            }
            // Render errors
            fieldErrors := ""
            if exists && len(fieldErrs) > 0 {
                var errorStrings []string
                for _, err := range fieldErrs {
                    errorStrings = append(errorStrings, template.HTMLEscapeString(err.Error()))
                }
                fieldErrors = fmt.Sprintf(`<div class="invalid-feedback d-block">%s</div>`, strings.Join(errorStrings, "<br>"))
            }
            // Wrap in form-check and bind label to input
            htmlStrings = append(htmlStrings, fmt.Sprintf(`<div class="mb-3">
<div class="form-check">
%s
<label class="form-check-label" for="%s">%s</label>
</div>
%s
</div>`, fieldHTML, name, label, fieldErrors))
            continue
        }

        // Add error styling to non-checkbox inputs if there are errors
        if exists && len(fieldErrs) > 0 {
            // Add Bootstrap invalid class to form controls
            fieldHTML = strings.ReplaceAll(fieldHTML, `class="form-control"`, `class="form-control is-invalid"`)
            fieldHTML = strings.ReplaceAll(fieldHTML, `class="form-select"`, `class="form-select is-invalid"`)
        } else {
            // Ensure form controls have proper Bootstrap classes
            if !strings.Contains(fieldHTML, `class="`) {
                fieldHTML = strings.ReplaceAll(fieldHTML, `<input`, `<input class="form-control"`)
                fieldHTML = strings.ReplaceAll(fieldHTML, `<select`, `<select class="form-select"`)
                fieldHTML = strings.ReplaceAll(fieldHTML, `<textarea`, `<textarea class="form-control"`)
            }
        }

		// Add Select2 and Flatpickr classes for enhanced controls
		if strings.Contains(fieldHTML, `<select`) && !strings.Contains(fieldHTML, `multiple`) {
			fieldHTML = strings.ReplaceAll(fieldHTML, `<select`, `<select data-role="select2"`)
		}
		if strings.Contains(fieldHTML, `type="date"`) {
			fieldHTML = strings.ReplaceAll(fieldHTML, `type="date"`, `type="text" data-role="flatpickr"`)
		}
		if strings.Contains(fieldHTML, `type="datetime-local"`) {
			fieldHTML = strings.ReplaceAll(fieldHTML, `type="datetime-local"`, `type="text" data-role="datetimepicker"`)
		}

		// Render field errors with Bootstrap styling
		fieldErrors := ""
		if exists && len(fieldErrs) > 0 {
			var errorStrings []string
			for _, err := range fieldErrs {
				errorStrings = append(errorStrings, template.HTMLEscapeString(err.Error()))
			}
			fieldErrors = fmt.Sprintf(`<div class="invalid-feedback d-block">%s</div>`, strings.Join(errorStrings, "<br>"))
		}

        htmlStrings = append(htmlStrings, fmt.Sprintf(`<div class="mb-3">
<label class="%s">%s</label>
%s
%s
</div>`, labelClass, label, fieldHTML, fieldErrors))
    }

	// Render form-level errors
	if len(formErrs) > 0 {
		var errorStrings []string
		for _, err := range formErrs {
			errorStrings = append(errorStrings, template.HTMLEscapeString(err.Error()))
		}
		formErrorsHTML := fmt.Sprintf(`<div class="alert alert-danger" role="alert">
<ul class="mb-0">%s</ul>
</div>`, "<li>"+strings.Join(errorStrings, "</li><li>")+"</li>")
		htmlStrings = append([]string{formErrorsHTML}, htmlStrings...)
	}

	return strings.Join(htmlStrings, "\n"), nil
}
