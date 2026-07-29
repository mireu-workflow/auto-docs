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

## 구현 기능 TODO

1. AI Studio 기반 Gemma4 모델 연동 및 에이전트 환경 구축
- [ ] **Google AI Studio API 연동**: AI Studio SDK 또는 REST API를 활용해 Gemma4 호출 클라이언트 구현
- [ ] **인증 처리**: `GEMINI_API_KEY` (또는 AI Studio API Key)를 환경 변수 및 CI/CD Secret을 통해 주입받는 보안 매커니즘 구현
- [ ] **MCP 클라이언트(Client) 구현**: Gemma4가 호출한 Tool Call(도구 실행 요청)을 가로채서 MCP 서버에 전달하고 결과를 돌려주는 오케스트레이터 개발

2. MCP (Model Context Protocol) 서버 구성 및 도구 제공
- [ ] **코드 분석 특화 MCP 서버 연동**:
    - **FileSystem MCP**: 소스코드를 읽고 쓸 수 있는 파일 시스템 접근 도구 연동
    - **Git MCP**: 변경사항(Diff), 커밋 로그 등을 확인하여 변경 분만 분석할 수 있는 도구 연동
- [ ] **커스텀 분석 MCP 서버 개발 (선택)**: 정적 분석(Linter), 빌드 로그 파싱 등 특정 프로젝트 분석을 위한 고유 도구 개발

3. 위키(Wiki) 기반 리포트 생성 및 자동 배포
- [ ] **마크다운 리포트 생성**: MCP를 통해 수집된 다각도 분석 정보를 구조화된 Markdown Wiki 포맷으로 렌더링
- [ ] **Github Wiki 자동 배포**: Github Wiki 저장소 자동 Git Clone -> 문서 커밋 -> Push 자동화
- [ ] **Gitlab Wiki 자동 배포**: Gitlab Wiki API 또는 Git 배포 기능을 통한 연동 자동화
- [ ] **Wiki 네비게이션 관리**: 신규 생성 및 업데이트된 리포트를 Wiki `_Sidebar.md` 또는 `Home.md`에 링크 자동 반영

4. CI/CD 환경 내 MCP 실행 자동화
- [ ] **백그라운드 MCP 프로세스 제어**: CI/CD 러너가 실행될 때 MCP 서버를 백그라운드로 구동하고, 실행 완료 후 프로세스를 안전하게 종료하는 제어 스크립트 작성
- [ ] **Github Actions 연동**: Docker 또는 Node.js 환경에서 MCP 서버와 분석 에이전트가 자동 동작하는 Workflow 개발
- [ ] **Gitlab Runner 및 Jenkins 연동**: 격리된 빌드 환경(Runner/Agent)을 위한 실행 스크립트(Makefile, Shell script 등) 개발

### 구현 도구

- Golang v1.26>
- Container 
- MCP 배포 방법 (미 계획)