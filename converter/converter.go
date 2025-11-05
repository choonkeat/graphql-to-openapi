package converter

import (
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// Config holds converter configuration
type Config struct {
	Title              string
	Version            string
	BaseURL            string
	PathPrefix         string
	DetectRESTPatterns bool
	CustomPlurals      map[string]string
	// Pluralization rules
	PluralizeSuffixesES  []string // Suffixes that get "es" added (e.g., s, x, z, ch, sh)
	PluralizeSuffixIES   string   // Suffix that triggers "ies" conversion (default "y")
	PluralizeDefaultSuffix string // Default suffix to add (default "s")
	// CRUD operation prefixes for REST pattern detection
	CRUDPrefixCreate string // Prefix for create operations (default "create")
	CRUDPrefixUpdate string // Prefix for update operations (default "update")
	CRUDPrefixDelete string // Prefix for delete operations (default "delete")
}

// Converter converts GraphQL schemas to OpenAPI
type Converter struct {
	config Config
	schema *ast.Schema
	doc    *OpenAPIDocument
}

// New creates a new converter
func New(config Config) *Converter {
	return &Converter{
		config: config,
	}
}

// Convert converts a GraphQL schema to OpenAPI
func (c *Converter) Convert(schemaSource string) (*OpenAPIDocument, error) {
	// Parse GraphQL schema
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Input: schemaSource,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL schema: %w", err)
	}

	c.schema = schema

	// Extract schema description (appears before the first type definition)
	schemaDesc := c.extractSchemaDescription(schemaSource)

	// Use first line as title if available, rest as description
	title := c.config.Title
	description := schemaDesc
	if schemaDesc != "" {
		lines := strings.Split(schemaDesc, "\n")
		firstLine := strings.TrimSpace(lines[0])
		if firstLine != "" && c.config.Title == "Converted from GraphQL" {
			// Use first line as title
			title = firstLine
			// Rest becomes description
			if len(lines) > 1 {
				description = strings.TrimSpace(strings.Join(lines[1:], "\n"))
			} else {
				description = ""
			}
		}
	}

	// Add footer to description
	footer := fmt.Sprintf("Converted from GraphQL (%s)", c.config.Version)
	if description != "" {
		description = description + "\n\n---\n\n" + footer
	} else {
		description = footer
	}

	c.doc = &OpenAPIDocument{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       title,
			Version:     c.config.Version,
			Description: description,
		},
		Paths:      make(map[string]*PathItem),
		Components: &Components{
			Schemas: make(map[string]*Schema),
		},
	}

	if c.config.BaseURL != "" {
		c.doc.Servers = []Server{{URL: c.config.BaseURL}}
	}

	// Detect REST patterns if enabled
	restPatterns := make(map[string]*RESTPattern)
	if c.config.DetectRESTPatterns {
		restPatterns = c.detectRESTPatterns()
	}

	// Convert types to schemas
	for _, typeDef := range schema.Types {
		if isBuiltInType(typeDef.Name) {
			continue
		}
		switch typeDef.Kind {
		case ast.Enum:
			c.convertEnumType(typeDef)
		case ast.Union:
			c.convertUnionType(typeDef)
		case ast.Interface:
			c.convertInterfaceType(typeDef)
		default:
			c.convertType(typeDef)
		}
	}

	// Convert queries and mutations
	if schema.Query != nil {
		c.convertQueries(schema.Query, restPatterns)
	}
	if schema.Mutation != nil {
		c.convertMutations(schema.Mutation, restPatterns)
	}
	if schema.Subscription != nil {
		c.convertSubscriptions(schema.Subscription)
	}

	return c.doc, nil
}

type RESTPattern struct {
	Resource   string // e.g., "user"
	Plural     string // e.g., "users"
	Type       *ast.Definition
	Operations map[string]bool // list, get, create, update, delete
}

func (c *Converter) detectRESTPatterns() map[string]*RESTPattern {
	patterns := make(map[string]*RESTPattern)

	// First pass: find list operations (e.g., users: [User!]!)
	if c.schema.Query != nil {
		for _, field := range c.schema.Query.Fields {
			if field.Type.Elem != nil && field.Type.Elem.NamedType != "" {
				// This is a list type
				typeName := field.Type.Elem.NamedType
				singular := c.singularize(field.Name)

				if singular != field.Name {
					// field.Name is plural
					if patterns[singular] == nil {
						patterns[singular] = &RESTPattern{
							Resource:   singular,
							Plural:     field.Name,
							Operations: make(map[string]bool),
						}
					}
					patterns[singular].Operations["list"] = true
					patterns[singular].Type = c.schema.Types[typeName]
				}
			}

			// Check for get by ID (e.g., user(id: ID!): User)
			if len(field.Arguments) == 1 && field.Arguments[0].Name == "id" {
				typeName := field.Type.Name()
				if field.Name == c.singularize(typeName) || strings.ToLower(field.Name) == strings.ToLower(typeName) {
					if patterns[field.Name] == nil {
						patterns[field.Name] = &RESTPattern{
							Resource:   field.Name,
							Plural:     c.pluralize(field.Name),
							Operations: make(map[string]bool),
						}
					}
					patterns[field.Name].Operations["get"] = true
					patterns[field.Name].Type = c.schema.Types[typeName]
				}
			}
		}
	}

	// Second pass: find mutations
	if c.schema.Mutation != nil {
		for _, field := range c.schema.Mutation.Fields {
			name := field.Name

			// Check for create{Resource}
			if c.config.CRUDPrefixCreate != "" && strings.HasPrefix(name, c.config.CRUDPrefixCreate) {
				resource := c.uncapitalize(strings.TrimPrefix(name, c.config.CRUDPrefixCreate))
				if patterns[resource] != nil {
					patterns[resource].Operations["create"] = true
				}
			}

			// Check for update{Resource}
			if c.config.CRUDPrefixUpdate != "" && strings.HasPrefix(name, c.config.CRUDPrefixUpdate) {
				resource := c.uncapitalize(strings.TrimPrefix(name, c.config.CRUDPrefixUpdate))
				if patterns[resource] != nil {
					patterns[resource].Operations["update"] = true
				}
			}

			// Check for delete{Resource}
			if c.config.CRUDPrefixDelete != "" && strings.HasPrefix(name, c.config.CRUDPrefixDelete) {
				resource := c.uncapitalize(strings.TrimPrefix(name, c.config.CRUDPrefixDelete))
				if patterns[resource] != nil {
					patterns[resource].Operations["delete"] = true
				}
			}
		}
	}

	// Filter: only keep patterns that have at least list + create
	filtered := make(map[string]*RESTPattern)
	for resource, pattern := range patterns {
		if pattern.Operations["list"] && pattern.Operations["create"] {
			filtered[resource] = pattern
			operationNames := []string{}
			for op := range pattern.Operations {
				operationNames = append(operationNames, op)
			}
			fmt.Printf("Detected REST pattern '%s': consolidated %d operations → /%s\n",
				resource, len(pattern.Operations), pattern.Plural)
		}
	}

	return filtered
}

func (c *Converter) convertEnumType(typeDef *ast.Definition) {
	enumValues := []string{}
	for _, val := range typeDef.EnumValues {
		enumValues = append(enumValues, val.Name)
	}

	schema := &Schema{
		Type: "string",
		Enum: enumValues,
	}

	if typeDef.Description != "" {
		schema.Description = typeDef.Description
	}

	c.doc.Components.Schemas[typeDef.Name] = schema
}

func (c *Converter) convertUnionType(typeDef *ast.Definition) {
	oneOf := []*Schema{}
	for _, t := range typeDef.Types {
		oneOf = append(oneOf, &Schema{
			Ref: "#/components/schemas/" + t,
		})
	}

	schema := &Schema{
		OneOf: oneOf,
	}

	if typeDef.Description != "" {
		schema.Description = typeDef.Description
	}

	c.doc.Components.Schemas[typeDef.Name] = schema
}

func (c *Converter) convertInterfaceType(typeDef *ast.Definition) {
	// For interfaces, we'll create a schema that accepts any of the implementing types
	// This is similar to unions but we could also just make it a generic object
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}

	if typeDef.Description != "" {
		schema.Description = typeDef.Description
	}

	// Add properties from the interface fields
	for _, field := range typeDef.Fields {
		propSchema := c.convertFieldType(field.Type)
		if field.Description != "" {
			propSchema.Description = c.addFieldNamePrefix(field.Name, field.Description)
		}
		schema.Properties[field.Name] = propSchema
	}

	c.doc.Components.Schemas[typeDef.Name] = schema
}

func (c *Converter) convertType(typeDef *ast.Definition) {
	if typeDef.Kind != ast.Object && typeDef.Kind != ast.InputObject {
		return
	}

	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
		Required:   []string{},
	}

	if typeDef.Description != "" {
		schema.Description = typeDef.Description
	}

	for _, field := range typeDef.Fields {
		propSchema := c.convertFieldType(field.Type)

		// Add human-friendly prefix to field description
		if field.Description != "" {
			propSchema.Description = c.addFieldNamePrefix(field.Name, field.Description)
		}

		// Handle deprecated directive
		deprecated := field.Directives.ForName("deprecated")
		if deprecated != nil {
			propSchema.Deprecated = true
			if reason := deprecated.Arguments.ForName("reason"); reason != nil {
				depReason := strings.Trim(reason.Value.String(), "\"")
				// Add field name prefix to deprecated reason
				fullDeprecation := c.camelToTitle(field.Name) + " - DEPRECATED: " + depReason
				if propSchema.Description != "" {
					propSchema.Description = fullDeprecation
				} else {
					propSchema.Description = fullDeprecation
				}
			}
		}

		// Handle constraint directive
		constraint := field.Directives.ForName("constraint")
		if constraint != nil {
			c.applyConstraints(propSchema, constraint)
		}

		// Handle specifiedBy directive on the field's type
		if fieldType := c.schema.Types[field.Type.Name()]; fieldType != nil {
			if specifiedBy := fieldType.Directives.ForName("specifiedBy"); specifiedBy != nil {
				if urlArg := specifiedBy.Arguments.ForName("url"); urlArg != nil {
					url := strings.Trim(urlArg.Value.String(), "\"")
					c.applySpecifiedBy(propSchema, url)
				}
			}
		}

		// Handle object/list references
		fieldTypeName := field.Type.Name()
		if field.Type.Elem != nil {
			// This is a list
			elemType := field.Type.Elem.NamedType
			if !isScalarType(elemType) && !isBuiltInType(elemType) {
				// List of objects - don't embed it, it becomes a sub-resource endpoint
				continue
			}
			// Scalar list - keep it as an array property (already converted by convertFieldType)
		} else if !isScalarType(fieldTypeName) && !isBuiltInType(fieldTypeName) {
			// This is an object reference - convert to ID
			propSchema = &Schema{
				Type:        "string",
				Description: fmt.Sprintf("Reference to %s.id - use GET /%s/{%sId}", fieldTypeName, c.pluralize(strings.ToLower(fieldTypeName)), field.Name),
			}
			field.Name = field.Name + "Id"
		}

		schema.Properties[field.Name] = propSchema

		if field.Type.NonNull {
			schema.Required = append(schema.Required, field.Name)
		}
	}

	c.doc.Components.Schemas[typeDef.Name] = schema
}

func (c *Converter) convertQueries(queryType *ast.Definition, restPatterns map[string]*RESTPattern) {
	processedFields := make(map[string]bool)

	// First, handle REST patterns
	for resource, pattern := range restPatterns {
		plural := pattern.Plural
		path := c.addPrefix("/" + plural)

		// List operation
		if pattern.Operations["list"] {
			if c.doc.Paths[path] == nil {
				c.doc.Paths[path] = &PathItem{}
			}
			c.doc.Paths[path].Get = &Operation{
				OperationID: "list" + c.capitalize(plural),
				Summary:     "List " + plural,
				Responses: map[string]*Response{
					"200": {
						Description: "Successful response",
						Content: map[string]*MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "array",
									Items: &Schema{
										Ref: "#/components/schemas/" + pattern.Type.Name,
									},
								},
							},
						},
					},
				},
			}
			processedFields[plural] = true
		}

		// Get by ID operation
		if pattern.Operations["get"] {
			idPath := c.addPrefix("/" + plural + "/{id}")
			if c.doc.Paths[idPath] == nil {
				c.doc.Paths[idPath] = &PathItem{}
			}
			c.doc.Paths[idPath].Get = &Operation{
				OperationID: "get" + c.capitalize(resource),
				Summary:     "Get " + resource + " by ID",
				Parameters: []*Parameter{
					{
						Name:     "id",
						In:       "path",
						Required: true,
						Schema:   &Schema{Type: "string"},
					},
				},
				Responses: map[string]*Response{
					"200": {
						Description: "Successful response",
						Content: map[string]*MediaType{
							"application/json": {
								Schema: &Schema{
									Ref: "#/components/schemas/" + pattern.Type.Name,
								},
							},
						},
					},
				},
			}
			processedFields[resource] = true
		}
	}

	// Then handle remaining queries
	for _, field := range queryType.Fields {
		if processedFields[field.Name] {
			continue
		}

		// Skip GraphQL introspection queries
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		path := c.addPrefix("/" + field.Name)
		operation := c.convertQueryField(field)

		if c.doc.Paths[path] == nil {
			c.doc.Paths[path] = &PathItem{}
		}
		c.doc.Paths[path].Get = operation
	}

	// Add sub-resource endpoints for list fields on types
	for _, typeDef := range c.schema.Types {
		if isBuiltInType(typeDef.Name) || typeDef.Kind != ast.Object {
			continue
		}

		for _, field := range typeDef.Fields {
			if field.Type.Elem != nil && field.Type.Elem.NamedType != "" {
				// Skip scalar arrays (e.g., [String!]!) - they stay as fields, not sub-resources
				if isScalarType(field.Type.Elem.NamedType) {
					continue
				}

				// This is a list field - create sub-resource endpoint
				resourceName := strings.ToLower(typeDef.Name)
				path := c.addPrefix("/" + c.pluralize(resourceName) + "/{id}/" + field.Name)

				if c.doc.Paths[path] == nil {
					c.doc.Paths[path] = &PathItem{}
				}

				c.doc.Paths[path].Get = &Operation{
					OperationID: "get" + typeDef.Name + c.capitalize(field.Name),
					Summary:     "Get " + field.Name + " by " + resourceName,
					Parameters: []*Parameter{
						{
							Name:     "id",
							In:       "path",
							Required: true,
							Schema:   &Schema{Type: "string"},
						},
					},
					Responses: map[string]*Response{
						"200": {
							Description: "Successful response",
							Content: map[string]*MediaType{
								"application/json": {
									Schema: &Schema{
										Type: "array",
										Items: &Schema{
											Ref: "#/components/schemas/" + field.Type.Elem.NamedType,
										},
									},
								},
							},
						},
					},
				}
			}
		}
	}
}

func (c *Converter) convertMutations(mutationType *ast.Definition, restPatterns map[string]*RESTPattern) {
	processedFields := make(map[string]bool)

	// First, handle REST patterns
	for resource, pattern := range restPatterns {
		plural := pattern.Plural

		// Create operation
		if pattern.Operations["create"] {
			path := c.addPrefix("/" + plural)
			if c.doc.Paths[path] == nil {
				c.doc.Paths[path] = &PathItem{}
			}

			// Find the create mutation field
			var createField *ast.FieldDefinition
			for _, field := range mutationType.Fields {
				if field.Name == c.config.CRUDPrefixCreate+c.capitalize(resource) {
					createField = field
					break
				}
			}

			if createField != nil {
				c.doc.Paths[path].Post = c.convertMutationField(createField, "Create "+resource)
				processedFields[createField.Name] = true
			}
		}

		// Update operation
		if pattern.Operations["update"] {
			path := c.addPrefix("/" + plural + "/{id}")
			if c.doc.Paths[path] == nil {
				c.doc.Paths[path] = &PathItem{}
			}

			// Find the update mutation field
			var updateField *ast.FieldDefinition
			for _, field := range mutationType.Fields {
				if field.Name == c.config.CRUDPrefixUpdate+c.capitalize(resource) {
					updateField = field
					break
				}
			}

			if updateField != nil {
				op := c.convertMutationField(updateField, "Update "+resource)
				// Add id path parameter
				op.Parameters = append([]*Parameter{
					{
						Name:     "id",
						In:       "path",
						Required: true,
						Schema:   &Schema{Type: "string"},
					},
				}, op.Parameters...)
				c.doc.Paths[path].Put = op
				processedFields[updateField.Name] = true
			}
		}

		// Delete operation
		if pattern.Operations["delete"] {
			path := c.addPrefix("/" + plural + "/{id}")
			if c.doc.Paths[path] == nil {
				c.doc.Paths[path] = &PathItem{}
			}

			// Find the delete mutation field
			var deleteField *ast.FieldDefinition
			for _, field := range mutationType.Fields {
				if field.Name == c.config.CRUDPrefixDelete+c.capitalize(resource) {
					deleteField = field
					break
				}
			}

			if deleteField != nil {
				op := c.convertMutationField(deleteField, "Delete "+resource)
				// For delete, id is usually a parameter
				if len(deleteField.Arguments) == 1 && deleteField.Arguments[0].Name == "id" {
					op.Parameters = []*Parameter{
						{
							Name:     "id",
							In:       "path",
							Required: true,
							Schema:   &Schema{Type: "string"},
						},
					}
					op.RequestBody = nil
				}
				c.doc.Paths[path].Delete = op
				processedFields[deleteField.Name] = true
			}
		}
	}

	// Then handle remaining mutations
	for _, field := range mutationType.Fields {
		if processedFields[field.Name] {
			continue
		}

		path := c.addPrefix("/" + field.Name)
		operation := c.convertMutationField(field, "")

		if c.doc.Paths[path] == nil {
			c.doc.Paths[path] = &PathItem{}
		}
		c.doc.Paths[path].Post = operation
	}
}

func (c *Converter) convertSubscriptions(subscriptionType *ast.Definition) {
	for _, field := range subscriptionType.Fields {
		// Skip GraphQL introspection fields
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		operation := c.convertSubscriptionField(field)
		path := c.buildSubscriptionPath(field)

		if c.doc.Paths[path] == nil {
			c.doc.Paths[path] = &PathItem{}
		}
		c.doc.Paths[path].Get = operation
	}
}

func (c *Converter) buildSubscriptionPath(field *ast.FieldDefinition) string {
	basePath := "/" + field.Name

	// Find the first required parameter to use in the path
	// This follows REST conventions where required params are in the path
	for _, arg := range field.Arguments {
		if arg.Type.NonNull {
			// First required parameter goes in the path
			return c.addPrefix(basePath + "/{" + arg.Name + "}")
		}
	}

	// No required parameters, just use the base path
	return c.addPrefix(basePath)
}

func (c *Converter) convertSubscriptionField(field *ast.FieldDefinition) *Operation {
	// Enhance description if needed
	enhancedDesc := c.addFieldNamePrefix(field.Name, field.Description)
	summary, description := c.splitDescription(enhancedDesc)

	// Get the return type name for the event type
	returnTypeName := field.Type.Name()
	if field.Type.Elem != nil {
		returnTypeName = field.Type.Elem.NamedType
	}

	// Build SSE format description
	sseDescription := fmt.Sprintf(`Server-Sent Events (SSE) stream.

Each event is formatted as:
  event: %s
  data: <JSON-encoded %s object>

Example:
  event: %s
  data: {"id":"123",...}

The connection remains open and events are pushed as they occur.
Use the EventSource API in browsers or any SSE client library.`, field.Name, returnTypeName, field.Name)

	if description != "" {
		sseDescription = description + "\n\n" + sseDescription
	}

	op := &Operation{
		OperationID: "subscribe" + c.capitalize(field.Name),
		Summary:     "Subscribe: " + summary,
		Description: sseDescription,
		Parameters:  []*Parameter{},
		Responses: map[string]*Response{
			"200": {
				Description: "SSE stream of " + returnTypeName + " events",
				Content: map[string]*MediaType{
					"text/event-stream": {
						Schema: &Schema{
							Type:        "string",
							Description: fmt.Sprintf("Server-Sent Events stream. Each event contains a %s object in JSON format.", returnTypeName),
						},
					},
				},
			},
		},
	}

	if op.Summary == "" || op.Summary == "Subscribe: " {
		op.Summary = "Subscribe to " + c.camelToTitle(field.Name)
	}

	// Handle deprecated directive
	deprecated := field.Directives.ForName("deprecated")
	if deprecated != nil {
		op.Deprecated = true
		if reason := deprecated.Arguments.ForName("reason"); reason != nil {
			depReason := strings.Trim(reason.Value.String(), "\"")
			op.Summary = "DEPRECATED: " + depReason
			if op.Description != "" {
				op.Description = "DEPRECATED: " + depReason + "\n\n" + op.Description
			} else {
				op.Description = "DEPRECATED: " + depReason
			}
		}
	}

	// Convert arguments to parameters
	// Required parameters become path parameters (handled in buildSubscriptionPath)
	// Optional parameters become query parameters
	pathParamUsed := false
	for _, arg := range field.Arguments {
		if arg.Type.NonNull && !pathParamUsed {
			// First required parameter goes in path
			param := &Parameter{
				Name:        arg.Name,
				In:          "path",
				Required:    true,
				Schema:      c.convertFieldType(arg.Type),
				Description: arg.Description,
			}
			op.Parameters = append(op.Parameters, param)
			pathParamUsed = true
		} else {
			// Other parameters go in query
			param := &Parameter{
				Name:        arg.Name,
				In:          "query",
				Required:    arg.Type.NonNull,
				Schema:      c.convertFieldType(arg.Type),
				Description: arg.Description,
			}

			// Handle array parameters with explode
			if arg.Type.Elem != nil {
				param.Style = "form"
				param.Explode = true
			}

			op.Parameters = append(op.Parameters, param)
		}
	}

	return op
}

func (c *Converter) convertQueryField(field *ast.FieldDefinition) *Operation {
	// Add human-friendly prefix if needed
	enhancedDesc := c.addFieldNamePrefix(field.Name, field.Description)
	summary, description := c.splitDescription(enhancedDesc)

	op := &Operation{
		OperationID: field.Name,
		Summary:     summary,
		Description: description,
		Parameters:  []*Parameter{},
		Responses: map[string]*Response{
			"200": {
				Description: "Successful response",
				Content: map[string]*MediaType{
					"application/json": {
						Schema: c.convertFieldType(field.Type),
					},
				},
			},
		},
	}

	if op.Summary == "" {
		op.Summary = c.camelToTitle(field.Name)
		op.Description = ""
	}

	// Handle deprecated directive
	deprecated := field.Directives.ForName("deprecated")
	if deprecated != nil {
		op.Deprecated = true
		if reason := deprecated.Arguments.ForName("reason"); reason != nil {
			depReason := strings.Trim(reason.Value.String(), "\"")
			op.Summary = "DEPRECATED: " + depReason
			if op.Description != "" {
				op.Description = "DEPRECATED: " + depReason + "\n\n" + op.Description
			} else {
				op.Description = "DEPRECATED: " + depReason
			}
		}
	}

	// Convert arguments to query parameters
	for _, arg := range field.Arguments {
		param := &Parameter{
			Name:     arg.Name,
			In:       "query",
			Required: arg.Type.NonNull,
			Schema:   c.convertFieldType(arg.Type),
		}

		if arg.Description != "" {
			param.Description = arg.Description
		}

		// Handle array parameters with explode
		if arg.Type.Elem != nil {
			param.Style = "form"
			param.Explode = true
		}

		op.Parameters = append(op.Parameters, param)
	}

	return op
}

func (c *Converter) convertMutationField(field *ast.FieldDefinition, fallbackSummary string) *Operation {
	var opSummary, opDescription string

	// Prefer GraphQL field description over generated summary
	if field.Description != "" {
		// Add human-friendly prefix if needed
		enhancedDesc := c.addFieldNamePrefix(field.Name, field.Description)
		opSummary, opDescription = c.splitDescription(enhancedDesc)
	} else if fallbackSummary != "" {
		opSummary = fallbackSummary
		opDescription = fallbackSummary
	} else {
		opSummary = c.camelToTitle(field.Name)
		opDescription = ""
	}

	op := &Operation{
		OperationID: field.Name,
		Summary:     opSummary,
		Description: opDescription,
		Parameters:  []*Parameter{},
		Responses: map[string]*Response{
			"200": {
				Description: "Successful response",
				Content: map[string]*MediaType{
					"application/json": {
						Schema: c.convertFieldType(field.Type),
					},
				},
			},
		},
	}

	// Handle deprecated directive
	deprecated := field.Directives.ForName("deprecated")
	if deprecated != nil {
		op.Deprecated = true
		if reason := deprecated.Arguments.ForName("reason"); reason != nil {
			depReason := strings.Trim(reason.Value.String(), "\"")
			op.Summary = "DEPRECATED: " + depReason
			if op.Description != "" {
				op.Description = "DEPRECATED: " + depReason + "\n\n" + op.Description
			} else {
				op.Description = "DEPRECATED: " + depReason
			}
		}
	}

	// Convert arguments to request body
	if len(field.Arguments) > 0 {
		bodySchema := &Schema{
			Type:       "object",
			Properties: make(map[string]*Schema),
			Required:   []string{},
		}

		for _, arg := range field.Arguments {
			propSchema := c.convertFieldType(arg.Type)
			if arg.Description != "" {
				propSchema.Description = arg.Description
			}
			bodySchema.Properties[arg.Name] = propSchema
			if arg.Type.NonNull {
				bodySchema.Required = append(bodySchema.Required, arg.Name)
			}
		}

		op.RequestBody = &RequestBody{
			Required: true,
			Content: map[string]*MediaType{
				"application/json": {
					Schema: bodySchema,
				},
			},
		}
	}

	return op
}

func (c *Converter) convertFieldType(fieldType *ast.Type) *Schema {
	// Handle lists
	if fieldType.Elem != nil {
		return &Schema{
			Type:  "array",
			Items: c.convertFieldType(fieldType.Elem),
		}
	}

	// Handle named types
	typeName := fieldType.NamedType

	switch typeName {
	case "Int":
		return &Schema{Type: "integer", Format: "int32"}
	case "Float":
		return &Schema{Type: "number", Format: "double"}
	case "String":
		return &Schema{Type: "string"}
	case "Boolean":
		return &Schema{Type: "boolean"}
	case "ID":
		return &Schema{Type: "string"}
	default:
		// Don't create references to built-in types like Query, Mutation, Subscription
		// These are GraphQL-specific and don't translate well to REST APIs
		if isBuiltInType(typeName) {
			return &Schema{Type: "object"}
		}

		// Reference to custom type
		if c.schema.Types[typeName] != nil {
			kind := c.schema.Types[typeName].Kind
			if kind == ast.Object || kind == ast.InputObject || kind == ast.Enum || kind == ast.Union || kind == ast.Interface {
				return &Schema{Ref: "#/components/schemas/" + typeName}
			}
		}
		// Fallback for custom scalars
		return &Schema{Type: "string"}
	}
}

func (c *Converter) applyConstraints(schema *Schema, directive *ast.Directive) {
	for _, arg := range directive.Arguments {
		raw := strings.Trim(arg.Value.Raw, "\"")
		switch arg.Name {
		case "minLength", "maxLength":
			if v := parseInt(raw); v != nil {
				if arg.Name == "minLength" {
					schema.MinLength = v
				} else {
					schema.MaxLength = v
				}
			}
		case "min", "max":
			if v := parseFloat(raw); v != nil {
				if arg.Name == "min" {
					schema.Minimum = v
				} else {
					schema.Maximum = v
				}
			}
		case "pattern":
			schema.Pattern = raw
		case "format":
			schema.Format = raw
		}
	}
}

func parseInt(s string) *int {
	if v, err := fmt.Sscanf(s, "%d", new(int)); err == nil && v == 1 {
		var result int
		fmt.Sscanf(s, "%d", &result)
		return &result
	}
	return nil
}

func parseFloat(s string) *float64 {
	if v, err := fmt.Sscanf(s, "%f", new(float64)); err == nil && v == 1 {
		var result float64
		fmt.Sscanf(s, "%f", &result)
		return &result
	}
	return nil
}

func (c *Converter) applySpecifiedBy(schema *Schema, url string) {
	// Try to infer format from common URLs
	if strings.Contains(url, "rfc4122") || strings.Contains(strings.ToLower(url), "uuid") {
		schema.Format = "uuid"
	} else if strings.Contains(strings.ToLower(url), "date-time") {
		schema.Format = "date-time"
	}

	if schema.Description != "" {
		schema.Description += "\n\nSpec: " + url
	} else {
		schema.Description = "Spec: " + url
	}
}

func (c *Converter) addPrefix(path string) string {
	if c.config.PathPrefix != "" {
		return c.config.PathPrefix + path
	}
	return path
}

func (c *Converter) pluralize(word string) string {
	// Check custom plurals (suffix match)
	for suffix, replacement := range c.config.CustomPlurals {
		if strings.HasSuffix(word, suffix) {
			return strings.TrimSuffix(word, suffix) + replacement
		}
	}

	// Check if word ends with any suffix that requires "es"
	for _, suffix := range c.config.PluralizeSuffixesES {
		if strings.HasSuffix(word, suffix) {
			return word + "es"
		}
	}

	// Check for "y" -> "ies" conversion
	if c.config.PluralizeSuffixIES != "" && strings.HasSuffix(word, c.config.PluralizeSuffixIES) &&
		len(word) > 1 && !isVowel(rune(word[len(word)-2])) {
		return word[:len(word)-1] + "ies"
	}

	// Default suffix
	return word + c.config.PluralizeDefaultSuffix
}

func (c *Converter) singularize(word string) string {
	// Check custom plurals in reverse
	for suffix, replacement := range c.config.CustomPlurals {
		if strings.HasSuffix(word, replacement) {
			return strings.TrimSuffix(word, replacement) + suffix
		}
	}

	// Reverse "ies" -> "y" conversion
	if c.config.PluralizeSuffixIES != "" && strings.HasSuffix(word, "ies") && len(word) > 3 {
		return word[:len(word)-3] + c.config.PluralizeSuffixIES
	}

	// Reverse "es" suffix for special endings
	if strings.HasSuffix(word, "es") && len(word) > 2 {
		base := word[:len(word)-2]
		for _, suffix := range c.config.PluralizeSuffixesES {
			if strings.HasSuffix(base, suffix) {
				return base
			}
		}
		// Not a special case, just remove "s"
		return word[:len(word)-1]
	}

	// Remove default suffix
	if strings.HasSuffix(word, c.config.PluralizeDefaultSuffix) && len(word) > 1 {
		return word[:len(word)-len(c.config.PluralizeDefaultSuffix)]
	}

	return word
}

func (c *Converter) capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (c *Converter) uncapitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func (c *Converter) extractSchemaDescription(schemaSource string) string {
	// Extract description from top of schema (before any type definitions)
	// Look for """ ... """ or # comments at the start
	lines := strings.Split(schemaSource, "\n")
	var desc []string
	inBlockComment := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for triple-quote block comments
		if strings.HasPrefix(trimmed, `"""`) {
			if inBlockComment {
				// End of block comment
				inBlockComment = false
				continue
			} else {
				// Start of block comment
				inBlockComment = true
				// Check if it's a single-line comment
				if strings.Count(trimmed, `"""`) >= 2 {
					// Single line: """content"""
					content := strings.TrimPrefix(trimmed, `"""`)
					content = strings.TrimSuffix(content, `"""`)
					desc = append(desc, strings.TrimSpace(content))
					inBlockComment = false
				}
				continue
			}
		}

		if inBlockComment {
			desc = append(desc, trimmed)
			continue
		}

		// Stop at first type definition
		if strings.HasPrefix(trimmed, "type ") ||
			strings.HasPrefix(trimmed, "interface ") ||
			strings.HasPrefix(trimmed, "scalar ") ||
			strings.HasPrefix(trimmed, "enum ") ||
			strings.HasPrefix(trimmed, "union ") ||
			strings.HasPrefix(trimmed, "input ") ||
			strings.HasPrefix(trimmed, "directive ") {
			break
		}

		// Skip empty lines at the end
		if trimmed == "" && len(desc) > 0 {
			break
		}
	}

	return strings.TrimSpace(strings.Join(desc, "\n"))
}

func (c *Converter) camelToTitle(s string) string {
	// Convert camelCase or PascalCase to Title Case
	// e.g., "addComment" -> "Add Comment", "emailAddress" -> "Email Address"
	if s == "" {
		return s
	}

	var result []rune
	runes := []rune(s)

	for i, r := range runes {
		if i == 0 {
			// Capitalize first letter
			result = append(result, []rune(strings.ToUpper(string(r)))...)
		} else if r >= 'A' && r <= 'Z' {
			// Add space before uppercase letters
			result = append(result, ' ')
			result = append(result, r)
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}

func (c *Converter) addFieldNamePrefix(fieldName string, description string) string {
	if description == "" {
		return c.camelToTitle(fieldName)
	}

	// Check if description already starts with a good action verb
	// These indicate the description is already well-formatted
	actionVerbs := []string{
		"List ", "Get ", "Fetch ", "Retrieve ", "Find ",
		"Create ", "Add ", "Insert ", "Post ",
		"Update ", "Modify ", "Edit ", "Put ", "Patch ",
		"Delete ", "Remove ", "Destroy ",
		"Search ", "Query ", "Filter ",
	}

	for _, verb := range actionVerbs {
		if strings.HasPrefix(description, verb) {
			// Description already starts with an action verb, keep as-is
			return description
		}
	}

	// Prepend human-friendly field name
	return c.camelToTitle(fieldName) + " - " + description
}

func (c *Converter) splitDescription(text string) (summary string, description string) {
	if text == "" {
		return "", ""
	}

	// Find first sentence-ending punctuation (. ! ? : ; -)
	punctuations := []string{". ", "! ", "? ", ": ", "; ", " - "}
	firstPunctIdx := -1
	firstPunctLen := 0

	for _, punc := range punctuations {
		idx := strings.Index(text, punc)
		if idx != -1 && (firstPunctIdx == -1 || idx < firstPunctIdx) {
			firstPunctIdx = idx
			firstPunctLen = len(punc)
		}
	}

	if firstPunctIdx == -1 {
		// No punctuation found, use entire text as summary
		return text, text
	}

	// Split at first punctuation
	// For " - ", don't include the dash in summary
	if firstPunctLen == 3 { // " - "
		summary = strings.TrimSpace(text[:firstPunctIdx])
	} else {
		summary = strings.TrimSpace(text[:firstPunctIdx+1])
	}
	description = strings.TrimSpace(text)

	return summary, description
}

func isVowel(r rune) bool {
	return r == 'a' || r == 'e' || r == 'i' || r == 'o' || r == 'u' ||
		r == 'A' || r == 'E' || r == 'I' || r == 'O' || r == 'U'
}

func isBuiltInType(name string) bool {
	return name == "Query" || name == "Mutation" || name == "Subscription" ||
		strings.HasPrefix(name, "__")
}

func isScalarType(name string) bool {
	return name == "Int" || name == "Float" || name == "String" || name == "Boolean" || name == "ID"
}
