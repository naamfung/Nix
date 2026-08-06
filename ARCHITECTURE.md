# DeepSeek-Reasonix 系統架構分析報告

## 1. 項目基礎信息

- **模塊名稱**: `inx`
- **Go 版本**: 1.25.0 (toolchain go1.26.5)
- **核心依賴**: 
  - `charm.land/bubbles/v2`, `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2` (TUI 界面庫)
  - `github.com/BurntSushi/toml` (配置解析)
  - `github.com/tree-sitter/*` (代碼語法樹解析)
  - `github.com/larksuite/oapi-sdk-go/v3` (飛書 API SDK)
  - `github.com/zalando/go-keyring` (密碼本管理)

## 2. 核心架構原則

根據 `INX.md` 的項目約定，系統遵循以下核心設計原則：

1. **Go 內核位於 `internal/`**: 每個包只負責一個明確的職責，並在包註釋中進行文檔說明。
2. **傳輸無關的控制器 (Transport-agnostic Controller)**: 一個 `control.Controller` 位於所有前端（聊天 TUI、HTTP/SSE 服務、Wails 桌面應用）之後。所有前端以相同的方式驅動 Controller（發送命令、渲染事件），且都不重新實現 turn lifecycle、cancellation 或 approval。
3. **Cache-first 策略**: system-prompt prefix（base prompt + tools + memory）必須在 turn 之間保持 byte-stable，以維持 DeepSeek 的自動 prefix cache 熱狀態。

## 3. 架構層次描述

系統採用分層架構設計，從上到下包括以下層次：

### 3.1 表現層 / 前端層 (Frontend / Presentation Layer)

- **CLI / TUI 接口 (`internal/cli/`)**: 
  - `cli.go`: 命令行入口，負責子命令路由、標誌解析、從配置組裝和退出代碼。支持 `run`, `chat`, `code`, `serve`, `setup`, `config`, `init`, `acp`, `mcp`, `remote`, `plugin`, `subagent`, `doctor`, `report`, `session`, `hook`, `task`, `review` 等子命令。
  - `chat_tui.go`: 基於 `bubbletea` 的終端用戶界面 (TUI) 核心實現（超過 16 萬行代碼）。
- **HTTP/SSE 服務端 (`internal/serve/`)**: 
  - 將 `control.Controller` 暴露為 HTTP 接口：將 typed event stream 作為 Server-Sent Events (SSE)，並將命令作為小 JSON POST 端點。是多個瀏覽器標籤頁共享單一會話的基礎。
- **多平台機器人適配 (`internal/bot/`)**: 
  - `gateway.go` 和相關適配器（`feishu/`, `qq/`, `weixin/`）支持多平台機器人適配（飛書、QQ、微信等），處理連接循環、控制服務器、渲染、會話管理、通道配置、路由配置、白名單配置等。
- **ACP 協議層 (`internal/acp/`)**: 
  - 實現 Agent-Client-Protocol，支持與 IDE/編輯器的深度集成，包含客戶端/服務端協議、會話管理、狀態管理等。

### 3.2 核心控制層 (Control / Orchestration Layer)

- **`internal/control/controller.go`**: 
  - 這是整個項目的核心樞紐，是一個 **transport-agnostic session driver**。
  - 擁有 agent run loop 和 session lifecycle，接收命令（Send/Cancel/Approve/SetPlanMode/Compact/NewSession/…），並將 reasoning、tool calls、approvals、turn completion 等作為 typed event stream 發送到單一的 `event.Sink`。
  - 管理會話恢復、記憶快照、MCP 工具/插件、技能 (skills)、目標 (goals)、工作區租約 (workspace lease)、檢查點 (checkpoints)、授權/詢問 (approval/ask) 等。

### 3.3 Agent 執行層 (Agent Execution Layer)

- **`internal/agent/agent.go`**: 
  - 核心執行組件，負責 tool calls、execution、provider interactions。
  - 包含 `DeliveryRuntimeMarker` 常量（delivery-first mode，強調 verifiable acceptance criteria 和 `complete_step` 的 evidence flow）。
  - 提供 `Renderer` 和 `Asker` 接口，以及 `callContextKey`、`parentSessionContextKey`、`subagentDepthContextKey`、`userImagesContextKey` 等 context key 用於傳遞上下文信息。
  - `Gate` 接口用於決定是否允許 tool call 執行。

### 3.4 模型提供商層 (Model Provider Layer)

- **`internal/provider/`**: 
  - 定義了 model-backend abstraction 和註冊表，映射 provider "kind" 到 factory。
  - 包含核心數據結構：`Message`, `ToolCall`, `ToolSchema`, `Request`。
  - 具體實現位於子包中（如 `provider/openai`, `provider/anthropic`, `provider/responses` 等），負責與不同 LLM 提供商（OpenAI、Anthropic、DeepSeek 等）進行交互，處理消息格式轉換、工具調用、圖片處理、會話清理（SanitizeToolPairing, ModelMessages）等。

### 3.5 工具和插件層 (Tool & Plugin Layer)

- **`internal/tool/`**: 
  - 定義了 `Tool` 接口（`Name`, `Description`, `Schema`, `Execute`, `ReadOnly`）。
  - 包含 `Previewer`（預覽文件更改）、`ImageTool`（返回帶圖片的工具結果）、`PlanModeClassifier`（聲明工具在規劃階段的執行立場）等接口。
  - 包含 MCP 相關接口：`MCPMetadata`, `MCPVisibleMetadata`, `MCPServerAuthorization`, `MCPAnnotations` 等。
  - 內置工具位於 `internal/tool/builtin/`。
- **`internal/plugin/`**: 
  - 負責處理 MCP（Model Context Protocol）服務器的啟動、連接（stdio, HTTP, SSE）、緩存、安全策略等。
  - 包含 `transport_stdio.go`, `transport_http.go`, `transport_sse.go` 等傳輸層實現，以及 `launcher_lock.go`, `security.go`, `lazy.go` 等核心邏輯。

### 3.6 配置與生態層 (Configuration & Ecosystem Layer)

- **`internal/config/`**: 負責加載和解析配置（TOML 格式），管理模型、提供者、工具、MCP 服務器、技能等配置。
- **`internal/memory/`**: 管理記憶快照、turn-tail notes 隊列、持久化機制。
- **`internal/skill/`**: 管理技能的發現、啟用子集、重新載入存儲等。
- **`internal/tool/builtin/`**: 內置工具實現（如文件操作、代碼搜索、shell 執行等）。

## 4. 數據流和控制流

- **數據流**: 用戶輸入 -> 前端（CLI/TUI、HTTP/SSE、Bot） -> `control.Controller` -> `agent.Agent` -> `provider.Provider` / `tool.Tool` / `plugin.MCP` -> 執行結果 -> `event.Sink` -> 前端渲染。
- **控制流**: `Controller` 驅動 `agent run loop` 和 `session lifecycle`，管理命令、會話、授權、計劃模式、記憶、技能、MCP 工具等；`Agent` 負責具體的執行邏輯（tool execution、provider requests、context management、evidence flow）。

## 5. 總結

**DeepSeek-Reasonix** 是一個高度模塊化、傳輸無關的 AI Agent 系統。其核心設計原則是「傳輸無關的控制器」（transport-agnostic Controller），位於 `internal/control/` 目錄，負責會話驅動、事件流管理和任務調度。所有前端（命令行 TUI、HTTP/SSE 服務、桌面應用、多平台機器人）都共享同一個控制器層，確保行為一致性。模型提供者層（`internal/provider/`）抽象了不同 LLM 提供商的接口，而工具和插件系統（`internal/tool/`, `internal/plugin/`）支持 MCP（Model Context Protocol）和內置工具。整個系統採用 Go 語言開發，遵循嚴格的包職責分離和緩存優先（cache-first）策略。
