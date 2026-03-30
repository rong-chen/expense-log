# 🪙 expense-log (端到端智能记账)

`expense-log` 是一款基于 **Golang** 构建的下一代智能记账后端。它彻底摒弃了传统的“OCR + 文本解析”的繁琐流程，采用 **多模态大模型 (Vision Language Model)** 直接对支付截图、发票和照片进行视觉理解，实现秒级的结构化账单生成。

---

## ✨ 核心特性

- **👁️ 视觉直达 (VLM)**: 不再依赖传统的 OCR 引擎。直接通过多模态模型（如 Qwen-VL, LLaVA 或 GPT-4o）理解图片内容。
- **🧠 语义深度理解**: 能够识别图片中的隐藏信息，例如通过商户 Logo 识别品牌，或从复杂的超市小票中提取汇总金额。
- **🛡️ 隐私安全**: 支持通过 **Ollama** 或 **LocalAI** 在本地运行视觉模型，财务数据无需出库。
- **⚡ 高并发异步流**: 基于 Golang 的协程池处理大尺寸图像流，支持任务状态实时追踪。

---

## 🛠️ 技术栈

| 模块 | 技术选型 | 描述 |
| :--- | :--- | :--- |
| **语言** | **Golang 1.24+** | 核心逻辑与并发模型 |
| **视觉模型** | **Qwen2-VL / LLaVA** | 部署于 Ollama，直接进行图像到 JSON 的转换 |
| **Web 框架** | **Gin / Echo** | 轻量级高性能 API 接口 |
| **数据库** | **PostgreSQL** | 存储结构化账单与流水 |
| **中间件** | **Redis (Asynq)** | 处理视觉模型推理的异步长耗时任务 |
| **存储** | **MinIO / Local** | 原始账单图片加密存储 |

---

## 🚀 核心架构：视觉提取流程

不同于传统方案，`expense-log` 简化了处理链路：

```mermaid
graph TD
    User[用户上传图片] --> API(Gin API)
    API --> Queue[Redis Task Queue]
    Queue --> Worker[Golang Worker]
    Worker --> VLM[本地多模态模型: Ollama/Qwen-VL]
    VLM -- 直接返回 JSON --> DB[(PostgreSQL)]
    DB --> Notify[前端状态更新]