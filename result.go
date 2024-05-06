package jsonschema

import "github.com/kaptinlin/go-i18n"

type EvaluationError struct {
	keyword string                 `json:"-"`
	code    string                 `json:"-"`
	message string                 `json:"-"`
	params  map[string]interface{} `json:"-"`
}

func NewEvaluationError(keyword string, code string, message string, params ...map[string]interface{}) *EvaluationError {
	if len(params) > 0 {
		return &EvaluationError{
			keyword: keyword,
			code:    code,
			message: message,
			params:  params[0],
		}
	} else {
		return &EvaluationError{
			keyword: keyword,
			code:    code,
			message: message,
		}
	}
}

func (e *EvaluationError) Error() string {
	return replace(e.message, e.params)
}

func (e *EvaluationError) Localize(localizer *i18n.Localizer) string {
	if localizer != nil {
		return localizer.Get(e.code, i18n.Vars(e.params))
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
	localizer        *i18n.Localizer             `json:"-"`
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

	e.CollectAnnotations()

	return e
}

func (e *EvaluationResult) SetEvaluationPath(evaluationPath string) *EvaluationResult {
	e.EvaluationPath = evaluationPath

	return e
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

func (e *EvaluationResult) SetLocalizer(localizer *i18n.Localizer) *EvaluationResult {
	e.localizer = localizer
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

	e.Errors[err.keyword] = err
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

	list := &List{
		Valid:            e.Valid,
		EvaluationPath:   e.EvaluationPath,
		SchemaLocation:   e.SchemaLocation,
		InstanceLocation: e.InstanceLocation,
		Annotations:      e.Annotations,
		Errors:           e.convertErrors(),
		Details:          make([]List, 0),
	}

	if hierarchyIncluded {
		for _, detail := range e.Details {
			childList := detail.ToList(true) // recursively include hierarchy
			list.Details = append(list.Details, *childList)
		}
	} else {
		e.flattenDetailsToList(list, e.Details) // flat structure
	}

	return list
}

func (e *EvaluationResult) flattenDetailsToList(list *List, details []*EvaluationResult) {
	for _, detail := range details {
		flatDetail := List{
			Valid:            detail.Valid,
			EvaluationPath:   detail.EvaluationPath,
			SchemaLocation:   detail.SchemaLocation,
			InstanceLocation: detail.InstanceLocation,
			Annotations:      detail.Annotations,
			Errors:           detail.convertErrors(),
		}
		list.Details = append(list.Details, flatDetail)

		if len(detail.Details) > 0 {
			e.flattenDetailsToList(list, detail.Details)
		}
	}
}

func (e *EvaluationResult) convertErrors() map[string]string {
	errors := make(map[string]string)
	for key, err := range e.Errors {
		if e.localizer != nil {
			errors[key] = err.Localize(e.localizer)
		} else {
			errors[key] = err.Error()
		}
	}
	return errors
}
