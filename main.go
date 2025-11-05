package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/choonkeat/graphql-to-openapi/converter"
)

func main() {
	var (
		schemaFile          = flag.String("schema", "", "GraphQL schema file (required)")
		outputFile          = flag.String("output", "openapi.yaml", "Output OpenAPI file")
		format              = flag.String("format", "yaml", "Output format: yaml or json")
		title               = flag.String("title", "Converted from GraphQL", "API title")
		version             = flag.String("version", "1.0.0", "API version")
		baseURL             = flag.String("base-url", "", "Base URL for the API")
		pathPrefix          = flag.String("path-prefix", "", "Path prefix for all endpoints (e.g., \"/api/v1\")")
		detectRESTPatterns  = flag.Bool("detect-rest-patterns", true, "Enable REST pattern detection")
		pluralizeSuffixes   = flag.String("pluralize-suffixes", "", "Custom pluralization suffix rules as JSON file")

		// Pluralization rules (advanced)
		pluralizeSuffixesES     = flag.String("pluralize-es-suffixes", "s,x,z,ch,sh", "Comma-separated suffixes that get 'es' added")
		pluralizeSuffixIES      = flag.String("pluralize-ies-suffix", "y", "Suffix that triggers 'ies' conversion")
		pluralizeDefaultSuffix  = flag.String("pluralize-default-suffix", "s", "Default suffix to add for pluralization")

		// CRUD prefixes (advanced)
		crudPrefixCreate = flag.String("crud-prefix-create", "create", "Prefix for create operations in REST pattern detection")
		crudPrefixUpdate = flag.String("crud-prefix-update", "update", "Prefix for update operations in REST pattern detection")
		crudPrefixDelete = flag.String("crud-prefix-delete", "delete", "Prefix for delete operations in REST pattern detection")

		help = flag.Bool("h", false, "Show help message")
	)
	flag.BoolVar(help, "help", false, "Show help message")
	flag.Parse()

	if *help || *schemaFile == "" {
		printHelp()
		if *schemaFile == "" {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Read schema file
	schemaBytes, err := os.ReadFile(*schemaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading schema file: %v\n", err)
		os.Exit(1)
	}

	// Load custom pluralization rules if provided
	var customPlurals map[string]string
	if *pluralizeSuffixes != "" {
		data, err := os.ReadFile(*pluralizeSuffixes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading pluralization file: %v\n", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(data, &customPlurals); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing pluralization JSON: %v\n", err)
			os.Exit(1)
		}
	}

	// Parse comma-separated ES suffixes
	var esSuffixes []string
	if *pluralizeSuffixesES != "" {
		for _, s := range strings.Split(*pluralizeSuffixesES, ",") {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				esSuffixes = append(esSuffixes, trimmed)
			}
		}
	}

	// Configure converter
	config := converter.Config{
		Title:              *title,
		Version:            *version,
		BaseURL:            *baseURL,
		PathPrefix:         *pathPrefix,
		DetectRESTPatterns: *detectRESTPatterns,
		CustomPlurals:      customPlurals,
		PluralizeSuffixesES:  esSuffixes,
		PluralizeSuffixIES:   *pluralizeSuffixIES,
		PluralizeDefaultSuffix: *pluralizeDefaultSuffix,
		CRUDPrefixCreate:     *crudPrefixCreate,
		CRUDPrefixUpdate:     *crudPrefixUpdate,
		CRUDPrefixDelete:     *crudPrefixDelete,
	}

	// Convert
	conv := converter.New(config)
	openAPIDoc, err := conv.Convert(string(schemaBytes))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting schema: %v\n", err)
		os.Exit(1)
	}

	// Output
	var output []byte
	if strings.ToLower(*format) == "json" {
		output, err = json.MarshalIndent(openAPIDoc, "", "  ")
	} else {
		output, err = converter.MarshalYAML(openAPIDoc)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling output: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputFile, output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted %s to %s\n", filepath.Base(*schemaFile), *outputFile)
}

func printHelp() {
	fmt.Print(`GraphQL to OpenAPI Converter

Usage: graphql-to-openapi [options]

Basic Options:
  -schema string
        GraphQL schema file (required)

  -output string
        Output OpenAPI file (default "openapi.yaml")

  -format string
        Output format: yaml or json (default "yaml")

API Metadata:
  -title string
        API title (default "Converted from GraphQL")

  -version string
        API version (default "1.0.0")

  -base-url string
        Base URL for the API

  -path-prefix string
        Path prefix for all endpoints (e.g., "/api/v1")

REST Pattern Detection:
  -detect-rest-patterns
        Enable REST pattern detection (default true)
        Detects CRUD patterns and consolidates them into REST endpoints

  -pluralize-suffixes string
        Custom pluralization suffix rules as JSON file
        Matches and replaces word endings (suffix match, not whole word)
        Example: {"person": "people", "child": "children", "data": "data"}

Advanced: Pluralization Rules
  -pluralize-es-suffixes string
        Comma-separated suffixes that get 'es' added (default "s,x,z,ch,sh")
        Example: "s,x,z,ch,sh" means "box" -> "boxes", "bus" -> "buses"

  -pluralize-ies-suffix string
        Suffix that triggers 'ies' conversion (default "y")
        Example: "y" means "category" -> "categories"

  -pluralize-default-suffix string
        Default suffix to add for pluralization (default "s")
        Example: "s" means "user" -> "users"

Advanced: CRUD Operation Prefixes
  -crud-prefix-create string
        Prefix for create operations in REST pattern detection (default "create")
        Example: "create" matches "createUser", "createPost"

  -crud-prefix-update string
        Prefix for update operations in REST pattern detection (default "update")
        Example: "update" matches "updateUser", "updatePost"

  -crud-prefix-delete string
        Prefix for delete operations in REST pattern detection (default "delete")
        Example: "delete" matches "deleteUser", "deletePost"

Help:
  -h, -help
        Show this help message

Examples:
  # Basic conversion
  graphql-to-openapi -schema schema.graphql -output api.yaml

  # With API metadata and versioning
  graphql-to-openapi -schema schema.graphql \
    -title "My API" \
    -version "2.0.0" \
    -path-prefix "/api/v2"

  # Disable REST pattern detection
  graphql-to-openapi -schema schema.graphql -detect-rest-patterns=false

  # Custom CRUD prefixes for non-English APIs
  graphql-to-openapi -schema schema.graphql \
    -crud-prefix-create "add" \
    -crud-prefix-update "modify" \
    -crud-prefix-delete "remove"

  # Custom pluralization for non-English words
  graphql-to-openapi -schema schema.graphql \
    -pluralize-default-suffix "en" \
    -pluralize-es-suffixes ""
`)
}
