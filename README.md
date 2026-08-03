<p align="center">
  <img src="docs/logo.svg" alt="Nix" width="640"/>
</p>

<p align="center">
  <strong>简体中文</strong>
  &nbsp;·&nbsp;
  <a href="./docs/GUIDE.md">指南</a>
  &nbsp;·&nbsp;
  <a href="./docs/ACP.md">ACP</a>
  &nbsp;·&nbsp;
  <a href="./docs/SPEC.md">規格</a>
  &nbsp;·&nbsp;
  <a href="https://github.com/naamfung/Nix/">官網</a>
</p>

<p align="center">
  <a href="https://www.npmjs.com/package/nix"><img src="https://img.shields.io/npm/v/nix.svg?style=flat-square&color=cb3837&labelColor=161b22&logo=npm&logoColor=white" alt="npm version"/></a>
  <a href="https://github.com/naamfung/Nix/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/naamfung/Nix/ci.yml?style=flat-square&label=ci&labelColor=161b22&logo=githubactions&logoColor=white" alt="CI"/></a>
  <a href="./LICENSE"><img src="https://img.shields.io/npm/l/nix.svg?style=flat-square&color=8b949e&labelColor=161b22" alt="license"/></a>
  <a href="https://www.npmjs.com/package/nix"><img src="https://img.shields.io/npm/dm/nix.svg?style=flat-square&color=3fb950&labelColor=161b22&label=downloads" alt="downloads"/></a>
  <a href="https://github.com/naamfung/Nix/stargazers"><img src="https://img.shields.io/github/stars/naamfung/Nix.svg?style=flat-square&color=dbab09&labelColor=161b22&logo=github&logoColor=white" alt="GitHub stars"/></a>
  <a href="https://atomgit.com/naamfung/Nix"><img src="https://atomgit.com/naamfung/Nix/star/badge.svg" alt="AtomGit stars"/></a>
</p>

<br/>

<h3 align="center">面向終端的 DeepSeek 原生 AI coding agent。</h3>
<p align="center">由配置與插件驅動的極薄 harness——單一靜態 Go 二進制，圍繞 DeepSeek 的前綴緩存調優，長會話也能把 token 成本壓低。</p>

<br/>

## 特性

- **配置驅動**：provider、agent、啟用的工具、插件全部在 `nix.toml` 中聲明，
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

選擇適合你的使用路徑。CLI/TUI、桌面端和 VS Code 擴展都使用同一套本地
Nix 引擎。

### 路徑 A：CLI / TUI

任意支持的平台都可以通過 npm 安裝原生二進制：

```sh
npm i -g nix                  # 任意系統;自動拉取對應平台的原生二進制
```

預編譯歸檔(`darwin|linux|windows × amd64|arm64`)和 `SHA256SUMS` 見每個
[GitHub release](https://github.com/naamfung/Nix/releases)。

### 路徑 B：桌面端

前往[官方下載頁](https://github.com/naamfung/Nix/releases)獲取最新桌面版本。

| 平台 | 包安裝包 | 架構 |
| --- | --- | --- |
| macOS | 通用 `.dmg` 或 `.zip` | Apple Silicon / Intel |
| Windows | 安裝器 `.exe` 或便攜 `.zip` | x64 / ARM64 |
| Linux | `.deb` 或 `.tar.gz` | x64 |

Windows 安裝器通過 [SignPath.io](https://signpath.io/) 完成代碼簽名，證書由
[SignPath 基金會](https://signpath.org/) 免費提供。

### 路徑 C：VS Code 擴展

請先完成路徑 A。擴展不內置 CLI，而是啟動本機的 `nix acp` 後端，
並提供原生聊天、編輯器上下文、工具調用審批、模型選擇和工作區會話。

- **VS Code：** [從 Visual Studio Marketplace 安裝](https://marketplace.visualstudio.com/items?itemName=SivanLiu.nix-agent)
- **VSCodium / Eclipse Theia：** [從 Open VSX Registry 安裝](https://open-vsx.org/extension/SivanLiu/nix-agent)
- **擴展 ID：** `SivanLiu.nix-agent` · [源碼與使用說明](https://github.com/SivanCola/nix-vscode)

### 路徑 D：從源碼構建

```sh
git clone https://github.com/naamfung/Nix.git
cd Nix
make build      # -> bin/nix(.exe)
make cross      # -> dist/（darwin|linux|windows × amd64|arm64）
```

## 快速開始

### CLI / TUI

以下命令僅適用於通過路徑 A 安裝的 CLI/TUI：

```sh
nix setup                      # 配置 provider 和模型
nix                            # 啟動交互式會話
nix run "把 main.go 裡的 TODO 實現掉"
```

需要項目指令時，可在交互式會話中運行 `/init`。

### 桌面端

從[官方下載頁](https://github.com/naamfung/Nix/releases)下載對應系統的安裝包，
安裝並啟動 Nix，然後在應用內配置 provider 和模型即可使用。桌面端無需執行
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
  <sub>This project is based on <strong>REASONIX</strong>. See the upstream repository at <a href="https://github.com/esengine/DeepSeek-Reasonix">esengine/DeepSeek-Reasonix</a>.</sub>
</p>
