# 🛠️ CLI 및 MCP Server 사용 가이드 (Usage Guide)

이 문서는 `auto-docs` CLI 도구의 실행 모드, MCP(Model Context Protocol) 서버 동작 방식, 및 상세 설정법을 다룹니다.

---

## 1. CLI 실행 모드

`auto-docs`는 크게 2가지 실행 모드를 지원합니다.

### A. 일반 자동 문서 생성 및 Wiki 배포 모드 (`./auto-docs`)
소스 코드 저장소의 Git 변경 내역(`git diff`) 및 주요 파일들을 수집하고, Gemma4 LLM을 통해 리포트를 작성한 후 지정된 GitHub/GitLab Wiki 저장소로 자동 커밋 & 푸시합니다.

```bash
./auto-docs
```

### B. 독립 MCP Server 모드 (`./auto-docs --mcp-server`)
Stdio 기반의 JSON-RPC 2.0 MCP(Model Context Protocol) 서버로 구동됩니다. 다른 MCP 클라이언트(Antigravity, Claude Desktop, Cursor 등)에서 `auto-docs`의 분석 도구를 가져다 쓸 수 있습니다.

```bash
./auto-docs --mcp-server
```

---

## 2. MCP (Model Context Protocol) 도구 제공 사양

`auto-docs`에 내장된 MCP Server는 공식 `mcp-go` SDK를 활용하며 다음과 같은 커스텀 도구를 제공합니다.

| 도구 이름 (Tool Name) | 설명 (Description) | 입력 파라미터 (Parameters) |
|---|---|---|
| `git_diff_analyzer` | 저장소의 커밋 간 Git diff 및 변경 파일 목록 추출 | `target` (string, 기본값: `"HEAD~1"`) |
| `filesystem_reader` | 분석 대상 소스 파일 내용을 안전하게 읽어 전달 | `files` (array of string) |

### MCP Server Config 예시 (`mcp_config.json`)

```json
{
  "mcpServers": {
    "auto-docs": {
      "command": "/path/to/auto-docs",
      "args": ["--mcp-server"],
      "env": {
        "REPO_PATH": "/path/to/target/repository"
      }
    }
  }
}
```

---

## 3. 환경 변수 상세 가이드

`auto-docs`는 환경 변수를 통해 동작을 제어합니다.

- `GEMINI_API_KEY` 또는 `AI_STUDIO_API_KEY`: Google AI Studio에서 발급받은 API Key입니다. (필수)
- `GEMINI_MODEL`: AI Studio 모델명 (기본값: `gemma-4`).
- `REPO_PATH`: 분석 대상 소스 코드 저장소 루트 경로 (기본값: `.`).
- `WIKI_REPO_URL`: 배포 대상 GitHub / GitLab Wiki 레포지토리 Git URL.
- `WIKI_TYPE`: Wiki 문법 및 배포 대상 구분 (`github` 또는 `gitlab`).
- `GIT_USERNAME`: Wiki 저장소 커밋 시 작성자 사용자명 (기본값: `auto-docs-bot`).
- `GIT_EMAIL`: Wiki 저장소 커밋 시 작성자 이메일 (기본값: `auto-docs@example.com`).

---

## 4. 빌드 및 컨테이너화 (Docker & Makefile)

### 로컬 빌드
```bash
make build
# 또는
go build -o auto-docs main.go
```

### Docker 이미지 빌드
```bash
docker build -t auto-docs:latest .
```

---

## 🔗 Related Pages
- [[Home]]
- [[GitHub Actions & GitLab Runner Setup Guide|GitHub-Actions-and-GitLab-Runner-Setup]]
