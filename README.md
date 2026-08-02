# auto-docs

## 구현 개요

어떤 씹새끼들이 나에게 일을 하면서 구현 명세서에다가 구라를 쳤다.
나는 그 구현 명세서가 구라인줄 모르고 이리 했다가 저리 했다가 IP 밴을 쳐먹었다.
그때의 나의 기분을 서술하시오.

이로 인해서 우리는 Gemma4를 활용하여 API 또는 기능들에 대해서 review를 하고 API검증까지는 못하겠지만.
~~어짜피 API검증은 우리가 할수있는 범위를 넘어셨음~~

그래서 소스코드를 보고 Report를 작성하는것이 우리의 목표이다.
그리고 그 Report 또한 Web으로 배포 하기 너무 비싸서, Github 및 Gitlab에 있는 Wiki 기능을 최대한 활용하고;
Github Actions, Gitlab Runner, 뭐 잘 되면 Jenkins에서도 돌아 갈수있도록 구현을 하고자 한다.

다른 Repo에서 Github Actions를 활용해서 적용이 가능하도록 구현 하여야한다.

## 구현 기능 TODO

1. AI Studio 기반 Gemma4 모델 연동 및 에이전트 환경 구축
- [x] **Google AI Studio API 연동**: AI Studio SDK / REST API 기반 Gemma4 호출 클라이언트 구현 (`pkg/agent`)
- [x] **인증 처리**: `GEMINI_API_KEY` (또는 AI Studio API Key) 환경 변수 및 CI/CD Secret 주입 보안 매커니즘 구현 (`pkg/config`)
- [x] **MCP 클라이언트(Client) 연동**: 공식 `mcp-go` SDK 기반 MCP Client 모듈 구현 (`pkg/mcpclient`)

2. MCP (Model Context Protocol) 서버 구성 및 도구 제공
- [x] **코드 분석 특화 MCP 서버 연동**:
    - **MCP Server Engine**: 공식 `mcp-go` SDK 기반 Stdio JSON-RPC 2.0 MCP 서버 구동 (`pkg/mcpserver`)
    - **FileSystem MCP Tool**: 소스코드를 읽고 전달하는 파일 시스템 도구 제공 (`pkg/mcpserver`, `pkg/analyzer`)
    - **Git MCP Tool**: 변경사항(Diff), 커밋 로그를 전달하는 Git 분석 도구 제공 (`pkg/mcpserver`, `pkg/analyzer`)

3. 위키(Wiki) 기반 리포트 생성 및 자동 배포
- [x] **마크다운 리포트 생성**: 수집된 코드 및 Git Diff 분석 정보를 Markdown Wiki 포맷으로 렌더링 (`pkg/agent`)
- [x] **Github Wiki 자동 배포**: Github Wiki 저장소 자동 Git Clone -> 문서 커밋 -> Push 자동화 (`pkg/wiki`)
- [x] **Gitlab Wiki 자동 배포**: Gitlab Wiki API 또는 Git 배포 기능을 통한 연동 자동화 (`pkg/wiki`)
- [x] **Wiki 네비게이션 관리**: 신규 생성 및 업데이트된 리포트를 Wiki `_Sidebar.md` 또는 `Home.md`에 링크 자동 반영 (`pkg/wiki`)

4. CI/CD 환경 내 MCP 실행 자동화
- [x] **백그라운드 MCP 프로세스 제어**: Stdio 및 독립 프로세스 기반 MCP 서버/클라이언트 실행 모드 지원 (`main.go --mcp-server`)
- [x] **Github Actions 연동**: Docker 및 Go 환경에서 CLI 분석 에이전트가 자동 동작하는 Action 및 Workflow 개발 (`action.yml`, `.github/workflows/auto-docs.yml`)
- [x] **Gitlab Runner 및 Jenkins 연동**: 격리된 빌드 환경(Runner/Agent)을 위한 실행 스크립트(`Makefile`, `Dockerfile`) 개발

### 구현 도구 및 컨테이너 MCP 사양

- **Golang**: v1.26+
- **Container / Docker**: Multi-stage 빌드 지원 (`Dockerfile`)
- **기본 탑재 MCP (Model Context Protocol) 도구**:
  - `git_diff_analyzer`: Git diff 및 변경 파일 목록 분석 도구 (`pkg/mcpserver`)
  - `filesystem_reader`: 소스코드 파일 내용 안전 조회 도구 (`pkg/mcpserver`)
- **MCP 실행 모드**:
  - `./auto-docs`: 일반 자동 보고서 생성 및 Wiki 자동 배포 모드
  - `./auto-docs --mcp-server`: Stdio 기반 JSON-RPC 2.0 독립 MCP 서버 구동 모드

### ⚙️ 주요 설정 및 환경 변수 (Environment Variables)

| 환경 변수 | 설명 | 기본값 |
| --- | --- | --- |
| `GEMINI_API_KEY` | Google AI Studio API 키 (필수) | - |
| `GEMINI_MODEL` | 사용할 Gemini/Gemma 모델 | `gemma-4` |
| `DOCS_LANGUAGE` | 생성될 문서 언어 (`ko`, `en`, `ja` 등) | `ko` |
| `GOAL_PROMPT` | LLM 페르소나 및 문서 작성 목표 프롬프트 | *"우리는 해당 코드를 보고 실행 여부를 떠나서, 현재 구현되어있는 코드를 기반으로 review하는 주니어 엔지니어이다. 우리는 구현되어있는 코드를 기반으로, 사용설명서 및 구현 명세서를 작성하여 wiki에다가 올려야하는 임무가 있다."* |
| `WIKI_REPO_URL` | 대상 Wiki 저장소 Git URL | CI 환경변수 자동 감지 |
| `WIKI_TYPE` | Wiki 저장소 타입 (`github` 또는 `gitlab`) | `github` |

### 📚 official Wiki & CI/CD Documentation
- 📖 [auto-docs GitLab Wiki 공식 사용설명서](https://gitlab.com/none-labs/gitworkflow-mcp/auto-docs/-/wikis)
- [GitHub Actions & GitLab Runner 셋업 가이드](wiki_docs/GitHub-Actions-and-GitLab-Runner-Setup.md)