#!/usr/bin/env bash
# Example: Self-Hosted Runner / Jenkins / Local Execution Script
set -euo pipefail

echo "🚀 Starting auto-docs Wiki Generator..."

# 1. Validate required environment variables
if [ -z "${GEMINI_API_KEY:-}" ]; then
    echo "❌ Error: GEMINI_API_KEY environment variable is missing."
    echo "Please export GEMINI_API_KEY='your-google-ai-studio-key'"
    exit 1
fi

# 2. Select execution runtime (Local binary or Docker container)
if command -v auto-docs &> /dev/null; then
    echo "📦 Executing installed auto-docs binary..."
    auto-docs -lang "${DOCS_LANGUAGE:-en}"
elif [ -f "./auto-docs" ]; then
    echo "📦 Executing ./auto-docs binary..."
    ./auto-docs -lang "${DOCS_LANGUAGE:-en}"
else
    echo "🐳 Executing auto-docs via Docker container..."
    docker run --rm \
      -e GEMINI_API_KEY="${GEMINI_API_KEY}" \
      -e GITHUB_TOKEN="${GITHUB_TOKEN:-}" \
      -e GITLAB_TOKEN="${GITLAB_TOKEN:-}" \
      -e WIKI_REPO_URL="${WIKI_REPO_URL:-}" \
      -e DOCS_LANGUAGE="${DOCS_LANGUAGE:-en}" \
      -v "$(pwd)":/app/repo \
      -e REPO_PATH="/app/repo" \
      ghcr.io/mireu-workflow/auto-docs:latest
fi

echo "✅ auto-docs execution completed successfully!"
