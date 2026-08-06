<h1 align="center">NIX</h1>

<p align="center">
  <strong>简体中文</strong>
  &nbsp;·&nbsp;
  <a href="./docs/GUIDE.md">指南</a>
  &nbsp;·&nbsp;
  <a href="./docs/ACP.md">ACP</a>
  &nbsp;·&nbsp;
  <a href="./docs/SPEC.md">規格</a>
  &nbsp;·&nbsp;
  <a href="https://github.com/naamfung/inx/">官網</a>
</p>

<p align="center">
  <a href="./LICENSE"><img src="https://img.shields.io/github/license/naamfung/inx.svg?style=flat-square&color=8b949e&labelColor=161b22" alt="license"/></a>
  <a href="https://github.com/naamfung/inx/stargazers"><img src="https://img.shields.io/github/stars/naamfung/inx.svg?style=flat-square&color=dbab09&labelColor=161b22&logo=github&logoColor=white" alt="GitHub stars"/></a>
</p>

<br/>

<h3 align="center">面向本地模型服務的原生 AI coding agent。</h3>
<p align="center">由配置與插件驅動的極薄 harness——單一靜態 Go 二進制，圍繞本地模型的前綴緩存調優，長會話也能把 token 成本壓低。</p>

<br/>

## 特性

- **配置驅動**：provider、agent、啟用的工具、插件全部在 `inx.toml` 中聲明，
  內核無硬編碼模型。
- **多模型 · 可組合**：DeepSeek 作為預設內建；任何 OpenAI 兼容
  端點都只是一條配置。可選讓兩個模型協同（執行器 + 規劃器），各自獨立、緩存穩定的 session。
- **插件驅動**：外部工具以子進程形式運行，通過 stdio JSON-RPC 通信（MCP 兼容）；
  內建工具在編譯期自註冊。
- **緩存友好的上下文維護**：啟動時注入穩定的環境摘要；舊工具輸出會先 snip/prune，
  再進入摘要 compaction；內建工具 schema 合約有文檔和回歸測試保護。
- **零摩擦分發**：`CGO_ENABLED=0` 單一二進制；一條命令交叉編譯到六個目標平台。
  唯一依賴是一個 TOML 解析庫。

## 安裝

### 方式一：下載預編譯版本（桌面端 / CLI）

預編譯歸檔(`darwin|linux|windows × amd64|arm64`)和 `SHA256SUMS` 見每個
[GitHub release](https://github.com/naamfung/inx/releases)。

#### 桌面端

前往[官方下載頁](https://github.com/naamfung/inx/releases)獲取最新桌面版本。

| 平台 | 包安裝包 | 架構 |
| --- | --- | --- |
| macOS | 通用 `.dmg` 或 `.zip` | Apple Silicon / Intel |
| Windows | 安裝器 `.exe` 或便攜 `.zip` | x64 / ARM64 |
| Linux | `.deb` 或 `.tar.gz` | x64 |

Windows 安裝器通過 [SignPath.io](https://signpath.io/) 完成代碼簽名，證書由
[SignPath 基金會](https://signpath.org/) 免費提供。

#### CLI / TUI

預編譯二進制歸檔可於 [GitHub release](https://github.com/naamfung/inx/releases) 下載。

### 方式二：從源碼構建

```sh
git clone https://github.com/naamfung/inx.git
cd Inx
make build      # -> bin/inx(.exe)
make cross      # -> dist/（darwin|linux|windows × amd64|arm64）
```

## 快速開始

### CLI / TUI

安裝後可使用以下命令：

```sh
inx setup                      # 配置 provider 和模型
inx                            # 啟動交互式會話
inx run "把 main.go 裡的 TODO 實現掉"
```

需要項目指令時，可在交互式會話中運行 `/init`。

### 桌面端

從[官方下載頁](https://github.com/naamfung/inx/releases)下載對應系統的安裝包，
安裝並啟動 Inx，然後在應用內配置 provider 和模型即可使用。桌面端無需執行
上面的 CLI 命令。

CLI 進階用法和詳細配置見 **[CLI 命令參考](./docs/CLI.md)**、
**[指南](./docs/GUIDE.md)** 和
**[配置路徑](./docs/CONFIG_PATHS.md)**。

## 文檔

- **開始使用：** [指南](./docs/GUIDE.md) ·
  [CLI 命令參考](./docs/CLI.md) · [配置路徑](./docs/CONFIG_PATHS.md) ·
  [ACP 編輯器接入](./docs/ACP.md)
- **功能與排障：** [子智能體 Profile](./docs/SUBAGENT_PROFILES.md) ·
  [Context Engine v2](./docs/SESSION_MEMORY_RETRIEVAL.md) ·
  [能力診斷](./docs/CAPABILITY_DIAGNOSTICS.md) ·
  [恢復與安全模式](./docs/RECOVERY.md) ·
  [機器人使用指南](./docs/BOT_GUIDE.md) ·
  [Checkpoints 與 rewind](./docs/CHECKPOINTS.md)
- **工程與遷移：** [規格](./docs/SPEC.md) ·
  [任務合約與暫停策略](./docs/TASK_CONTRACT.md) ·
  [工具合約](./docs/TOOL_CONTRACT.md) ·
  [從 0.x 遷移](./docs/MIGRATING.md)

---

<p align="center">
  <sub>MIT —— 見 <a href="./LICENSE">LICENSE</a></sub>
</p>

---

<p align="center">
  <sub>This project is based on <strong>INX</strong>. See the upstream repository at <a href="https://github.com/esengine/DeepSeek-Inx">esengine/DeepSeek-Inx</a>.</sub>
</p>
