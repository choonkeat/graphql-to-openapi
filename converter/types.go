package converter

// OpenAPIDocument represents an OpenAPI 3.0 document
type OpenAPIDocument struct {
	OpenAPI    string                `json:"openapi" yaml:"openapi"`
	Info       Info                  `json:"info" yaml:"info"`
	Servers    []Server              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths      map[string]*PathItem  `json:"paths" yaml:"paths"`
	Components *Components           `json:"components,omitempty" yaml:"components,omitempty"`
}

// Info contains API metadata
type Info struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string `json:"version" yaml:"version"`
}

// Server represents an API server
type Server struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// PathItem describes operations available on a path
type PathItem struct {
	Get     *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Post    *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Put     *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
	Options *Operation `json:"options,omitempty" yaml:"options,omitempty"`
}

// Operation describes a single API operation
type Operation struct {
	OperationID string              `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Summary     string              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string              `json:"description,omitempty" yaml:"description,omitempty"`
	Parameters  []*Parameter        `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]*Response `json:"responses" yaml:"responses"`
	Deprecated  bool                `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

// Parameter describes a single operation parameter
type Parameter struct {
	Name        string  `json:"name" yaml:"name"`
	In          string  `json:"in" yaml:"in"` // query, path, header, cookie
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool    `json:"required,omitempty" yaml:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
	Style       string  `json:"style,omitempty" yaml:"style,omitempty"`
	Explode     bool    `json:"explode,omitempty" yaml:"explode,omitempty"`
}

// RequestBody describes a request body
type RequestBody struct {
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool                  `json:"required,omitempty" yaml:"required,omitempty"`
	Content     map[string]*MediaType `json:"content" yaml:"content"`
}

// Response describes a single response
type Response struct {
	Description string                `json:"description" yaml:"description"`
	Content     map[string]*MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// MediaType describes a media type
type MediaType struct {
	Schema *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// Components holds reusable objects
type Components struct {
	Schemas map[string]*Schema `json:"schemas,omitempty" yaml:"schemas,omitempty"`
}

// Schema describes a data type
type Schema struct {
	Type        string             `json:"type,omitempty" yaml:"type,omitempty"`
	Format      string             `json:"format,omitempty" yaml:"format,omitempty"`
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Required    []string           `json:"required,omitempty" yaml:"required,omitempty"`
	Items       *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	Ref         string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Deprecated  bool               `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Enum        []string           `json:"enum,omitempty" yaml:"enum,omitempty"`
	OneOf       []*Schema          `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	MinLength   *int               `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength   *int               `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Minimum     *float64           `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum     *float64           `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	Pattern     string             `json:"pattern,omitempty" yaml:"pattern,omitempty"`
}
