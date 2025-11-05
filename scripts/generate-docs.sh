#!/bin/bash
# Script to generate separate GraphQL schema and OpenAPI documentation pages

set -e

DOCS_DIR="examples"
EXAMPLES_DIR="examples"
FAILED_EXAMPLES=()

for example_dir in "$EXAMPLES_DIR"/*/ ; do
    if [ ! -f "$example_dir/schema.graphql" ]; then
        continue
    fi

    name=$(basename "$example_dir")
    schema_file="$example_dir/schema.graphql"
    openapi_file="$example_dir/openapi.yaml"
    schema_html="$DOCS_DIR/$name.schema.html"
    yaml_html="$DOCS_DIR/$name.yaml.html"
    temp_redoc="$DOCS_DIR/$name.tmp.redoc.html"

    echo "Generating docs for $name..."

    # Generate ReDoc HTML using @redocly/cli
    if ! npx @redocly/cli build-docs "$openapi_file" -o "$temp_redoc" 2>&1; then
        echo "❌ Failed to generate ReDoc for $name"
        FAILED_EXAMPLES+=("$name")
        continue
    fi

    # Read and escape schema content for JavaScript
    schema_content=$(cat "$schema_file" | sed 's/\\/\\\\/g; s/`/\\`/g; s/\$/\\$/g')

    # Read and escape YAML content for JavaScript
    yaml_content=$(cat "$openapi_file" | sed 's/\\/\\\\/g; s/`/\\`/g; s/\$/\\$/g')

    # Create standalone GraphQL schema page
    cat > "$schema_html" <<EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>$name - GraphQL Schema</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism-tomorrow.min.css">
    <style id="custom-overrides">
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #1e1e1e;
            height: 100vh;
            overflow: auto;
        }
        pre[class*="language-"] {
            margin: 0 !important;
            padding: 2rem !important;
            background: #1e1e1e !important;
            font-size: 0.9rem !important;
            line-height: 1.6 !important;
            min-height: 100vh;
        }
        code[class*="language-"] {
            font-family: 'Monaco', 'Menlo', 'Consolas', monospace !important;
        }
        /* Tone down string colors - make them less attention-grabbing */
        .token.string {
            color: #606060 !important;
        }
        /* Also tone down comments */
        .token.comment {
            color: #5a5a5a !important;
        }
    </style>
</head>
<body>
    <pre><code class="language-graphql" id="schema-code"></code></pre>

    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/prism.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-graphql.min.js"></script>
    <script>
        // Set schema content
        const schemaCode = \`$schema_content\`;
        document.getElementById('schema-code').textContent = schemaCode;

        // Apply syntax highlighting
        Prism.highlightAll();
    </script>
</body>
</html>
EOF

    # Create standalone OpenAPI YAML page
    cat > "$yaml_html" <<EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>$name - OpenAPI YAML</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism-tomorrow.min.css">
    <style id="custom-overrides">
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #1e1e1e;
            height: 100vh;
            overflow: auto;
        }
        pre[class*="language-"] {
            margin: 0 !important;
            padding: 2rem !important;
            background: #1e1e1e !important;
            font-size: 0.9rem !important;
            line-height: 1.6 !important;
            min-height: 100vh;
        }
        code[class*="language-"] {
            font-family: 'Monaco', 'Menlo', 'Consolas', monospace !important;
        }
        /* Tone down string colors - make them less attention-grabbing */
        .token.string {
            color: #606060 !important;
        }
        /* Also tone down comments */
        .token.comment {
            color: #5a5a5a !important;
        }
    </style>
</head>
<body>
    <pre><code class="language-yaml" id="yaml-code"></code></pre>

    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/prism.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-yaml.min.js"></script>
    <script>
        // Set YAML content
        const yamlCode = \`$yaml_content\`;
        document.getElementById('yaml-code').textContent = yamlCode;

        // Apply syntax highlighting
        Prism.highlightAll();
    </script>
</body>
</html>
EOF

    # Rename the ReDoc file
    mv "$temp_redoc" "$DOCS_DIR/$name.redoc.html"

    echo "✓ Generated $schema_html + $name.yaml.html + $name.redoc.html"
done

echo ""
if [ ${#FAILED_EXAMPLES[@]} -eq 0 ]; then
    echo "Documentation generated successfully!"
else
    echo "❌ Documentation generation completed with failures:"
    for failed in "${FAILED_EXAMPLES[@]}"; do
        echo "  - $failed"
    done
    exit 1
fi
