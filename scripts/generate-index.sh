#!/bin/bash
# Script to generate index page for documentation

DOCS_DIR="examples"
EXAMPLES_DIR="examples"

# Function to extract JSON value (simple parser for our use case)
get_json_value() {
    local json="$1"
    local key="$2"
    echo "$json" | grep -o "\"$key\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | sed 's/.*"\([^"]*\)"$/\1/'
}

# Generate example cards by reading metadata.json from each example directory
generate_example_cards() {
    for dir in "$EXAMPLES_DIR"/*/ ; do
        if [ -f "$dir/metadata.json" ]; then
            metadata=$(cat "$dir/metadata.json")
            number=$(get_json_value "$metadata" "number")
            title=$(get_json_value "$metadata" "title")
            description=$(get_json_value "$metadata" "description")
            dirname=$(basename "$dir")

            cat <<EOF
                <div class="example-card">
                    <h3 class="example-title">$title</h3>
                    <p class="example-description">
                        $description
                    </p>
                    <div class="example-links">
                        <a href="$dirname.schema.html" class="example-link schema-link">
                            <span class="link-icon">📝</span>
                            <span>GraphQL Schema</span>
                        </a>
                        <a href="$dirname.yaml.html" class="example-link yaml-link">
                            <span class="link-icon">📄</span>
                            <span>OpenAPI YAML</span>
                        </a>
                        <a href="$dirname.redoc.html" class="example-link docs-link">
                            <span class="link-icon">📖</span>
                            <span>OpenAPI Docs</span>
                        </a>
                    </div>
                </div>
EOF
        fi
    done
}

# Generate the main HTML file
cat > "$DOCS_DIR/index.html" <<'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GraphQL to OpenAPI Converter - Documentation</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            color: #333;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 2rem;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 3rem 2rem;
            text-align: center;
        }
        h1 {
            font-size: 2.5rem;
            margin-bottom: 0.5rem;
        }
        .tagline {
            font-size: 1.2rem;
            opacity: 0.9;
        }
        .content {
            padding: 3rem 2rem;
        }
        .intro {
            text-align: center;
            margin-bottom: 3rem;
            font-size: 1.1rem;
            color: #666;
        }
        .examples {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 2rem;
            margin-top: 2rem;
        }
        .example-card {
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            padding: 1.5rem;
            transition: all 0.3s ease;
            background: white;
            display: flex;
            flex-direction: column;
        }
        .example-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 30px rgba(102, 126, 234, 0.2);
            border-color: #667eea;
        }
        .example-title {
            font-size: 1.5rem;
            margin-bottom: 0.5rem;
            color: #333;
        }
        .example-description {
            color: #666;
            margin-bottom: 1.5rem;
            flex: 1;
        }
        .example-links {
            display: flex;
            gap: 0.75rem;
            flex-wrap: wrap;
        }
        .example-link {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 0.75rem 1rem;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 500;
            font-size: 0.85rem;
            transition: all 0.3s ease;
            flex: 1;
            min-width: 120px;
            justify-content: center;
        }
        .example-link:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
        }
        .link-icon {
            font-size: 1.1rem;
        }
        footer {
            text-align: center;
            padding: 2rem;
            color: #666;
            border-top: 1px solid #e0e0e0;
        }
        .features {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1.5rem;
            margin: 3rem 0;
        }
        @media (min-width: 1200px) {
            .features {
                grid-template-columns: repeat(5, 1fr);
            }
        }
        .feature {
            text-align: center;
            padding: 1.5rem;
            background: #f8f9fa;
            border-radius: 8px;
        }
        .feature-icon {
            font-size: 2rem;
            margin-bottom: 0.5rem;
        }
        .feature-title {
            font-weight: 600;
            margin-bottom: 0.5rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>GraphQL to OpenAPI Converter</h1>
            <p class="tagline">GraphQL the Good Parts 🤝 OpenAPI the Good Parts</p>
        </header>

        <div class="content">
            <div class="intro">
                <p>This tool converts GraphQL schemas to OpenAPI 3.0 specifications, eliminating N+1 queries and depth attacks by design.</p>
            </div>

            <div class="features">
                <div class="feature">
                    <div class="feature-icon">✅</div>
                    <div class="feature-title">No N+1 Queries</div>
                    <div>Explicit endpoints, no arbitrary nesting</div>
                </div>
                <div class="feature">
                    <div class="feature-icon">🛡️</div>
                    <div class="feature-title">No Depth Attacks</div>
                    <div>Flat response shapes by design</div>
                </div>
                <div class="feature">
                    <div class="feature-icon">⚡</div>
                    <div class="feature-title">HTTP Cacheable</div>
                    <div>Standard GET requests, predictable URLs</div>
                </div>
                <div class="feature">
                    <div class="feature-icon">🔒</div>
                    <div class="feature-title">Type Safe</div>
                    <div>Inherits GraphQL's type system</div>
                </div>
                <div class="feature">
                    <div class="feature-icon">✨</div>
                    <div class="feature-title">Domain-Focused</div>
                    <div>GraphQL schema is more succinct and meaningful than OpenAPI YAML</div>
                </div>
            </div>

            <h2 style="margin-top: 3rem; margin-bottom: 1.5rem; font-size: 2rem; text-align: center;">Examples</h2>

            <div class="examples">
EOF

# Insert dynamically generated example cards
generate_example_cards >> "$DOCS_DIR/index.html"

# Complete the HTML file
cat >> "$DOCS_DIR/index.html" <<'EOF'
            </div>
        </div>

        <footer>
            <p>Built with Go • Powered by gqlparser and Redoc</p>
            <p style="margin-top: 0.5rem; font-size: 0.9rem;">
                <a href="https://github.com/choonkeat/graphql-to-openapi" style="color: #667eea; text-decoration: none;">View on GitHub</a>
            </p>
        </footer>
    </div>
</body>
</html>
EOF

echo "✓ Generated index.html with $(ls -1d "$EXAMPLES_DIR"/*/ | wc -l | tr -d ' ') examples"
