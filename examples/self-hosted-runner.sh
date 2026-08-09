#!/usr/bin/env bash
# 예시 3: Self-Hosted Runner / Jenkins / 로컬 실행 스크립트
set -e

echo "🚀 Starting auto-docs Wiki Generator..."

# 1. 필수 환경변수 검증
if [ -z "$GEMINI_API_KEY" ]; then
    echo "❌ Error: GEMINI_API_KEY environment variable is missing."
    echo "Please export GEMINI_API_KEY='your-google-ai-studio-key'"
    exit 1
fi

# 2. 실행 환경 선택 (로컬 바이너리 또는 Docker)
if command -v auto-docs &> /dev/null; then
    echo "📦 Executing local auto-docs binary..."
    auto-docs
elif [ -f "./auto-docs" ]; then
    echo "📦 Executing ./auto-docs binary..."
    ./auto-docs
else
    echo "🐳 Executing auto-docs via Docker container..."
    docker run --rm \
      -e GEMINI_API_KEY="${GEMINI_API_KEY}" \
      -e GITHUB_TOKEN="${GITHUB_TOKEN}" \
      -e GITLAB_TOKEN="${GITLAB_TOKEN}" \
      -e WIKI_REPO_URL="${WIKI_REPO_URL}" \
      -v "$(pwd)":/app/repo \
      -e REPO_PATH="/app/repo" \
      registry.gitlab.com/none-labs/gitworkflow-mcp/auto-docs:latest
fi

echo "✅ auto-docs execution completed successfully!"
