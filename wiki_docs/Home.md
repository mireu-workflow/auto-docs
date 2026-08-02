# 🚀 auto-docs 사용설명 가이드 (Usage Guide)

`auto-docs`는 **Google AI Studio (Gemma4)** LLM 모델 및 **MCP (Model Context Protocol)** 기술을 활용하여, 코드베이스 변경 사항(Git Diff)과 소스 코드를 자동으로 분석하고 기술 문서(Technical Report)를 작성하여 GitHub Wiki 및 GitLab Wiki에 자동 배포해주는 커스텀 AI 문서화 에이전트입니다.

---

## 📌 주요 특징 (Key Features)

1. **Gemma4 AI Agent 기반 기술 문서 자동 작성**:
   - Git commit history, git diff, 소스 코드 구조를 수집하여 AI가 체계적인 기능 설명서 및 기술 리포트를 자동 생성합니다.
2. **공식 `mcp-go` 기반 MCP Server & Client 모듈 통합**:
   - `git_diff_analyzer` 및 `filesystem_reader` 도구를 Stdio JSON-RPC 2.0 기반으로 제공합니다.
3. **GitHub Wiki / GitLab Wiki 자동 배포**:
   - 생성된 리포트를 Target Wiki 저장소로 자동 Clone/Commit/Push 합니다.
   - Wiki `_Sidebar.md` 및 `Home.md` 목차에 신규 리포트 링크를 자동으로 갱신합니다.
4. **CI/CD 환경 완전 지원**:
   - GitHub Actions Composite Action (`Mireu-Lab/auto-docs@v1`) 및 Standalone Workflow 지원
   - GitLab Runner (`.gitlab-ci.yml`), Docker Multi-stage, Jenkins 지원

---

## 📖 Wiki 문서 목차 (Navigation)

- 📘 **[[GitHub Actions & GitLab Runner CI/CD 셋업 가이드|GitHub-Actions-and-GitLab-Runner-Setup]]**
  - GitHub Actions 연동 방법 (Composite Action / Direct Go Workflow)
  - GitLab Runner 연동 방법 (`.gitlab-ci.yml`, Docker-in-Docker)
  - Jenkins Pipeline 연동 방법
- 🛠️ **[[CLI 및 MCP Server 사용 가이드|CLI-and-MCP-Usage-Guide]]**
  - 로컬 환경 실행 및 CLI 파라미터 안내
  - Stdio 기반 MCP Server (`--mcp-server`) 동작 모드 및 도구 사양
  - 주요 환경 변수 설정 안내

---

## ⚡ 빠른 시작 (Quick Start)

### 1. 환경 변수 설정

| 환경 변수 | 설명 | 기본값 |
|---|---|---|
| `GEMINI_API_KEY` | Google AI Studio API 키 | (필수) |
| `GEMINI_MODEL` | 사용 모델 (예: `gemma-4`) | `gemma-4` |
| `WIKI_REPO_URL` | Target Wiki Git URL (OAuth/Token 포함) | N/A |
| `WIKI_TYPE` | Wiki 종류 (`github` 또는 `gitlab`) | `github` |
| `REPO_PATH` | 분석 대상 소스 코드 로컬 경로 | `.` |
| `GIT_USERNAME` | Wiki 커밋용 Git 사용자명 | `auto-docs-bot` |
| `GIT_EMAIL` | Wiki 커밋용 Git 이메일 | `auto-docs@example.com` |

### 2. 로컬 실행 Command

```bash
# Go 전용 실행
export GEMINI_API_KEY="your-gemini-api-key"
export WIKI_REPO_URL="https://oauth2:YOUR_TOKEN@gitlab.com/your-org/your-repo.wiki.git"
export WIKI_TYPE="gitlab"

go run main.go
```

```bash
# Docker 컨테이너 실행
docker run --rm \
  -e GEMINI_API_KEY="your-gemini-api-key" \
  -e WIKI_REPO_URL="https://oauth2:YOUR_TOKEN@gitlab.com/your-org/your-repo.wiki.git" \
  -e WIKI_TYPE="gitlab" \
  -v $(pwd):/app/repo \
  auto-docs:latest
```

---

## 🏗️ 시스템 아키텍처 (Architecture)

```mermaid
graph TD
    A[Git Repository Changes] -->|git diff / source inspect| B[auto-docs CLI Engine]
    B -->|Stdio JSON-RPC 2.0| C[MCP Server Engine]
    C -->|git_diff_analyzer| D[Git Diff Collector]
    C -->|filesystem_reader| E[Source Code Reader]
    D --> B
    E --> B
    B -->|Gemini API Prompt| F[Google AI Studio / Gemma4 LLM]
    F -->|Generated Markdown Report| B
    B -->|Git Clone & Push| G[GitHub / GitLab Wiki]
```

---

## 🤝 문의 및 기여 (Contributing & Support)

문제가 발생하거나 기능 제안이 있으신 경우 Project Merge Request 및 Issue를 등록해 주세요!
- Repository: [GitLab auto-docs](https://gitlab.com/none-labs/gitworkflow-mcp/auto-docs)
