# 🚀 auto-docs

Google AI Studio Gemma 4 모델 및 MCP(Model Context Protocol)를 활용하여 소스코드와 Git Diff를 분석하고, GitHub/GitLab Wiki에 기술 보고서 및 시스템 사용설명서를 자동 생성/배포하는 CI/CD 자동화 솔루션입니다.

---

## 💡 프로젝트 핵심 특징

- **Runner & Actions 토큰 기반 자동 배포**: `GITHUB_TOKEN` (GitHub Actions) 또는 `CI_JOB_TOKEN` / `PROJECT_ACCESS_TOKEN` / `GITLAB_TOKEN` (GitLab Runner)을 감지하여 별도의 복잡한 URL 구성 없이 저장소 Wiki에 바로 배포합니다.
- **Gemma 4 & Gemini AI 기반 문서화**: 최신 Gemma 4 / Gemini 모델을 연동하여 아키텍처 구조도(Mermaid), API/MCP 명세, 실행 방법, 환경 변수 표를 정교한 Markdown Wiki 포맷으로 작성합니다.
- **Stdio MCP Server 지원**: `./auto-docs --mcp-server` 플래그로 실행 시 JSON-RPC 2.0 기반 Stdio MCP Server로 동작하여 파일 읽기, 정규식 소스 검색, Git Diff 분석, DuckDuckGo 라이브 웹 검색 도구를 제공합니다.

---

## 🛠️ 구현 기능 Checklist

1. **AI Studio 기반 Gemma4 모델 연동 및 에이전트 환경 구축**
   - [x] **Google AI Studio API 연동**: REST API 기반 Gemma4 호출 클라이언트 구현 ([`pkg/agent`](file:///home/limmireu1214/_projects/GitWorkflow-MCP/auto-docs/pkg/agent))
   - [x] **보안 인증 처리**: `GEMINI_API_KEY` (또는 `AI_STUDIO_API_KEY`) 환경 변수 및 CI/CD Secret 주입 ([`pkg/config`](file:///home/limmireu1214/_projects/GitWorkflow-MCP/auto-docs/pkg/config))

2. **MCP (Model Context Protocol) 서버 구성 및 도구 제공**
   - [x] **코드 분석 특화 MCP 서버**: Stdio JSON-RPC 2.0 MCP 서버 구동 ([`pkg/mcpserver`](file:///home/limmireu1214/_projects/GitWorkflow-MCP/auto-docs/pkg/mcpserver))
   - [x] **FileSystem & Git Tools**: 파일 정밀 조회(`read_file`), 코드 검색(`code_search`), Git Diff 분석 도구 제공

3. **위키(Wiki) 기반 리포트 생성 및 자동 배포**
   - [x] **마크다운 리포트 생성**: 분석 정보를 Markdown Wiki 포맷으로 자동 생성 ([`pkg/agent`](file:///home/limmireu1214/_projects/GitWorkflow-MCP/auto-docs/pkg/agent))
   - [x] **GitHub / GitLab Wiki Token 기반 자동 배포**: CI/CD Runner 토큰으로 Wiki Git 저장소 Clone -> Commit -> Push 자동화 ([`pkg/wiki`](file:///home/limmireu1214/_projects/GitWorkflow-MCP/auto-docs/pkg/wiki))
   - [x] **Wiki 네비게이션 관리**: `Home.md` 및 `_Sidebar.md` 네비게이션 링크 자동 반영

4. **CI/CD 환경 내 실행 자동화**
   - [x] **GitHub Actions 연동**: Composite Action ([`action.yml`](file:///home/limmireu1214/_projects/GitWorkflow-MCP/auto-docs/action.yml)) 및 Workflow ([`.github/workflows/auto-docs.yml`](file:///home/limmireu1214/_projects/GitWorkflow-MCP/auto-docs/.github/workflows/auto-docs.yml))
   - [x] **GitLab Runner 연동**: [`.gitlab-ci.yml`](file:///home/limmireu1214/_projects/GitWorkflow-MCP/auto-docs/.gitlab-ci.yml) 및 Multi-stage Docker 이미지 ([`Dockerfile`](file:///home/limmireu1214/_projects/GitWorkflow-MCP/auto-docs/Dockerfile))

---

## ⚙️ 주요 설정 및 환경 변수 (Environment Variables)

| 환경 변수 | 설명 | 기본값 |
| --- | --- | --- |
| `GEMINI_API_KEY` | Google AI Studio API 키 (필수) | - |
| `GITHUB_TOKEN` | GitHub Actions 자동 주입 토큰 (GitHub Wiki 자동 배포용) | `${{ github.token }}` |
| `GITLAB_TOKEN` / `PROJECT_ACCESS_TOKEN` | GitLab Project/Personal Access Token (GitLab Wiki Push 권한) | - |
| `GEMINI_MODEL` | 사용할 Gemini/Gemma 모델 | `gemma-4-26b-a4b-it` |
| `DOCS_LANGUAGE` | 생성될 문서 언어 (`ko`, `en`, `ja` 등) | `ko` |
| `WIKI_REPO_URL` | Target Wiki Git Repository URL | Runner 환경 변수 기반 자동 감지 |
| `WIKI_TYPE` | Wiki 저장소 타입 (`github` 또는 `gitlab`) | `github` |

---

## 📚 Official Wiki Documentation

- 📖 [auto-docs GitLab Wiki 공식 사용설명서](https://gitlab.com/none-labs/gitworkflow-mcp/auto-docs/-/wikis)


---

## 📂 예시 Runner 및 Action 셋업 파일 (Example Setup Files)

- 🐙 [GitHub Actions Workflow 예시 (`examples/github-action-workflow.yml`)](examples/github-action-workflow.yml)
- 🦊 [GitLab Runner Pipeline 예시 (`examples/gitlab-ci-runner.yml`)](examples/gitlab-ci-runner.yml)
- 🖥️ [Self-Hosted Runner / Jenkins 스크립트 예시 (`examples/self-hosted-runner.sh`)](examples/self-hosted-runner.sh)