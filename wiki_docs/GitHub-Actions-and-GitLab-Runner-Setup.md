# 🚀 auto-docs CI/CD Setup Guide

Welcome to the **auto-docs** setup guide! This document explains how to integrate `auto-docs` into **GitHub Actions** and **GitLab Runner** to automatically generate technical reports and deploy them to your repository's Wiki using Gemma4 AI Agent and MCP (Model Context Protocol).

---

## 🐙 1. GitHub Actions Setup Guide

You can integrate `auto-docs` into your GitHub repository in two ways: using the **GitHub Composite Action** (`Mireu-Lab/auto-docs@v1`) or a **Go/Docker Workflow**.

### Prerequisites
1. **GitHub Wiki Enabled**: Go to Repository Settings -> Features -> Check **Wikis**. (Initialize at least one page like `Home.md` via UI if it's new).
2. **Gemini API Key Secret**: Go to Repository Settings -> Secrets and variables -> Actions -> Add `GEMINI_API_KEY`.

### Option A: Using Composite Action (Recommended for external repositories)

Create `.github/workflows/auto-docs.yml` in your repository:

```yaml
name: Auto Docs Wiki Generator

on:
  push:
    branches: [ main, master ]
  workflow_dispatch:

jobs:
  generate-docs:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      wiki: write

    steps:
      - name: Checkout Source Code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Run auto-docs Action
        uses: Mireu-Lab/auto-docs@v1
        with:
          gemini_api_key: ${{ secrets.GEMINI_API_KEY }}
          gemini_model: "gemma-4"
          wiki_repo_url: "https://x-access-token:${{ secrets.GITHUB_TOKEN }}@github.com/${{ github.repository }}.wiki.git"
          wiki_type: "github"
```

### Option B: Standalone Workflow (Direct Go Build)

```yaml
name: Auto Docs Direct Generator

on:
  push:
    branches: [ main, master ]

jobs:
  build-and-deploy-wiki:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go 1.26
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'

      - name: Run auto-docs
        env:
          GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
          GEMINI_MODEL: "gemma-4"
          WIKI_REPO_URL: "https://x-access-token:${{ secrets.GITHUB_TOKEN }}@github.com/${{ github.repository }}.wiki.git"
          GIT_USERNAME: "github-actions[bot]"
          GIT_EMAIL: "github-actions[bot]@users.noreply.github.com"
        run: |
          go run main.go
```

---

## 🦊 2. GitLab Runner Setup Guide

For GitLab projects, `auto-docs` supports containerized execution via `Dockerfile` or pre-built binaries.

### Prerequisites
1. **GitLab Wiki Enabled**: Ensure Wiki feature is enabled in Project Settings -> General -> Visibility, project features, permissions.
2. **Masked CI/CD Variables**: Go to Settings -> CI/CD -> Variables:
   - `GEMINI_API_KEY`: Google AI Studio API Key (Masked & Protected)
   - `GITLAB_WIKI_TOKEN`: Project Access Token or Personal Access Token with `write_repository` permission.

### `.gitlab-ci.yml` Configuration

```yaml
stages:
  - test
  - build-container
  - generate-docs

variables:
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: "/certs"
  IMAGE_TAG: $CI_REGISTRY_IMAGE:$CI_COMMIT_REF_SLUG
  IMAGE_TAG_LATEST: $CI_REGISTRY_IMAGE:latest

# 1. Go 소스코드 유닛 테스트
test:
  stage: test
  image: golang:1.26-alpine
  script:
    - apk add --no-cache git build-base
    - go test -v ./...

# 2. Docker 컨테이너 빌드 및 GitLab Container Registry 배포
build-container:
  stage: build-container
  image: docker:24.0.5
  services:
    - name: docker:24.0.5-dind
  before_script:
    - docker login -u "$CI_REGISTRY_USER" -p "$CI_REGISTRY_PASSWORD" "$CI_REGISTRY"
  script:
    - 'echo "Building Docker image: $IMAGE_TAG"'
    - 'docker build -t "$IMAGE_TAG" -t "$IMAGE_TAG_LATEST" .'
    - 'echo "Pushing Docker image to Container Registry..."'
    - 'docker push "$IMAGE_TAG"'
    - 'docker push "$IMAGE_TAG_LATEST"'
  only:
    - main
    - master
    - tags

# 3. auto-docs 실행 및 Wiki 자동 업데이트
generate-wiki-docs:
  stage: generate-docs
  image: docker:24.0.5
  services:
    - name: docker:24.0.5-dind
  rules:
    - if: '$GEMINI_API_KEY != null && ($CI_COMMIT_BRANCH == "main" || $CI_COMMIT_BRANCH == "master")'
      when: manual
      allow_failure: true
  script:
    - docker run --rm \
        -e GEMINI_API_KEY="${GEMINI_API_KEY}" \
        -e GEMINI_MODEL="gemma-4" \
        -e WIKI_REPO_URL="https://oauth2:${GITLAB_WIKI_TOKEN}@${CI_SERVER_HOST}/${CI_PROJECT_PATH}.wiki.git" \
        -e WIKI_TYPE="gitlab" \
        -e GIT_USERNAME="gitlab-ci-bot" \
        -e GIT_EMAIL="gitlab-ci@example.com" \
        -v $(pwd):/app/repo \
        -e REPO_PATH="/app/repo" \
        "$IMAGE_TAG_LATEST"
```

---

## 🚢 3. Jenkins Setup Guide

If you use Jenkins pipelines, you can run `auto-docs` using `Makefile` or Docker:

```groovy
pipeline {
    agent { docker { image 'golang:1.26-alpine' } }
    environment {
        GEMINI_API_KEY = credentials('gemini-api-key')
        WIKI_REPO_URL  = 'https://github.com/your-org/your-repo.wiki.git'
    }
    stages {
        stage('Generate Docs') {
            steps {
                sh 'apk add --no-cache git'
                sh 'go run main.go'
            }
        }
    }
}
```

---

## 🔗 Related Pages
- [[Home]]
- [[Architecture & MCP Specification|Home#mcp-model-context-protocol-도구-제공]]
