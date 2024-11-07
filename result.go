package jsonschema

import "github.com/kaptinlin/go-i18n"

type EvaluationError struct {
	Keyword string                 `json:"keyword"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Params  map[string]interface{} `json:"params"`
}

func NewEvaluationError(keyword string, code string, message string, params ...map[string]interface{}) *EvaluationError {
	if len(params) > 0 {
		return &EvaluationError{
			Keyword: keyword,
			Code:    code,
			Message: message,
			Params:  params[0],
		}
	} else {
		return &EvaluationError{
			Keyword: keyword,
			Code:    code,
			Message: message,
		}
	}
}

func (e *EvaluationError) Error() string {
	return replace(e.Message, e.Params)
}

func (e *EvaluationError) Localize(localizer *i18n.Localizer) string {
	if localizer != nil {
		return localizer.Get(e.Code, i18n.Vars(e.Params))
	} else {
		return e.Error()
	}
}

type Flag struct {
	Valid bool `json:"valid"`
}

type List struct {
	Valid            bool                   `json:"valid"`
	EvaluationPath   string                 `json:"evaluationPath"`
	SchemaLocation   string                 `json:"schemaLocation"`
	InstanceLocation string                 `json:"instanceLocation"`
	Annotations      map[string]interface{} `json:"annotations,omitempty"`
	Errors           map[string]string      `json:"errors,omitempty"`
	Details          []List                 `json:"details,omitempty"`
}

type EvaluationResult struct {
	schema           *Schema                     `json:"-"`
	Valid            bool                        `json:"valid"`
	EvaluationPath   string                      `json:"evaluationPath"`
	SchemaLocation   string                      `json:"schemaLocation"`
	InstanceLocation string                      `json:"instanceLocation"`
	Annotations      map[string]interface{}      `json:"annotations,omitempty"`
	Errors           map[string]*EvaluationError `json:"errors,omitempty"` // Store error messages here
	Details          []*EvaluationResult         `json:"details,omitempty"`
}

func NewEvaluationResult(schema *Schema) *EvaluationResult {
	e := &EvaluationResult{
		schema: schema,
		Valid:  true,
	}
	//nolint:errcheck
	e.CollectAnnotations()

	return e
}

func (e *EvaluationResult) SetEvaluationPath(evaluationPath string) *EvaluationResult {
	e.EvaluationPath = evaluationPath

	return e
}

func (e *EvaluationResult) Error() string {
	return "evaluation failed"
}

func (e *EvaluationResult) SetSchemaLocation(location string) *EvaluationResult {
	e.SchemaLocation = location

	return e
}

func (e *EvaluationResult) SetInstanceLocation(instanceLocation string) *EvaluationResult {
	e.InstanceLocation = instanceLocation

	return e
}

func (e *EvaluationResult) SetInvalid() *EvaluationResult {
	e.Valid = false

	return e
}

func (e *EvaluationResult) IsValid() bool {
	return e.Valid
}

func (e *EvaluationResult) AddError(err *EvaluationError) *EvaluationResult {
	if e.Errors == nil {
		e.Errors = make(map[string]*EvaluationError)
	}

	if e.Valid {
		e.Valid = false
	}

	e.Errors[err.Keyword] = err
	return e
}

func (e *EvaluationResult) AddDetail(detail *EvaluationResult) *EvaluationResult {
	if e.Details == nil {
		e.Details = make([]*EvaluationResult, 0)
	}

	e.Details = append(e.Details, detail)
	return e
}

func (e *EvaluationResult) AddAnnotation(keyword string, annotation interface{}) *EvaluationResult {
	if e.Annotations == nil {
		e.Annotations = make(map[string]interface{})
	}

	e.Annotations[keyword] = annotation
	return e
}

func (e *EvaluationResult) CollectAnnotations() *EvaluationResult {
	if e.Annotations == nil {
		e.Annotations = make(map[string]interface{})
	}

	if e.schema.Title != nil {
		e.Annotations["title"] = e.schema.Title
	}
	if e.schema.Description != nil {
		e.Annotations["description"] = e.schema.Description
	}
	if e.schema.Default != nil {
		e.Annotations["default"] = e.schema.Default
	}
	if e.schema.Deprecated != nil {
		e.Annotations["deprecated"] = e.schema.Deprecated
	}
	if e.schema.ReadOnly != nil {
		e.Annotations["readOnly"] = e.schema.ReadOnly
	}
	if e.schema.WriteOnly != nil {
		e.Annotations["writeOnly"] = e.schema.WriteOnly
	}
	if e.schema.Examples != nil {
		e.Annotations["examples"] = e.schema.Examples
	}

	return e
}

// Converts EvaluationResult to a simple Flag struct
func (e *EvaluationResult) ToFlag() *Flag {
	return &Flag{
		Valid: e.Valid,
	}
}

// ToList converts the evaluation results into a list format with optional hierarchy
// includeHierarchy is variadic; if not provided, it defaults to true
func (e *EvaluationResult) ToList(includeHierarchy ...bool) *List {
	// Set default value for includeHierarchy to true
	hierarchyIncluded := true
	if len(includeHierarchy) > 0 {
		hierarchyIncluded = includeHierarchy[0]
	}

	return e.ToLocalizeList(nil, hierarchyIncluded)
}

// ToLocalizeList converts the evaluation results into a list format with optional hierarchy with localization
// includeHierarchy is variadic; if not provided, it defaults to true
func (e *EvaluationResult) ToLocalizeList(localizer *i18n.Localizer, includeHierarchy ...bool) *List {
	// Set default value for includeHierarchy to true
	hierarchyIncluded := true
	if len(includeHierarchy) > 0 {
		hierarchyIncluded = includeHierarchy[0]
	}

	list := &List{
		Valid:            e.Valid,
		EvaluationPath:   e.EvaluationPath,
		SchemaLocation:   e.SchemaLocation,
		InstanceLocation: e.InstanceLocation,
		Annotations:      e.Annotations,
		Errors:           e.convertErrors(localizer),
		Details:          make([]List, 0),
	}

	if hierarchyIncluded {
		for _, detail := range e.Details {
			childList := detail.ToLocalizeList(localizer, true) // recursively include hierarchy
			list.Details = append(list.Details, *childList)
		}
	} else {
		e.flattenDetailsToList(localizer, list, e.Details) // flat structure
	}

	return list
}

func (e *EvaluationResult) flattenDetailsToList(localizer *i18n.Localizer, list *List, details []*EvaluationResult) {
	for _, detail := range details {
		flatDetail := List{
			Valid:            detail.Valid,
			EvaluationPath:   detail.EvaluationPath,
			SchemaLocation:   detail.SchemaLocation,
			InstanceLocation: detail.InstanceLocation,
			Annotations:      detail.Annotations,
			Errors:           detail.convertErrors(localizer),
		}
		list.Details = append(list.Details, flatDetail)

		if len(detail.Details) > 0 {
			e.flattenDetailsToList(localizer, list, detail.Details)
		}
	}
}

func (e *EvaluationResult) convertErrors(localizer *i18n.Localizer) map[string]string {
	errors := make(map[string]string)
	for key, err := range e.Errors {
		if localizer != nil {
			errors[key] = err.Localize(localizer)
		} else {
			errors[key] = err.Error()
		}
	}
	return errors
}
