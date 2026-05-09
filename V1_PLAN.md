# 客服系统 V1 版本规划文档

> **目标**：实现 P0 级别核心功能，搭建系统基础框架，完成 MVP 最小可用版本。
>
> **技术栈**：Go (后端) + Vue 3 + TypeScript (前端) + WebSocket (实时通信) + PostgreSQL + Redis + Kafka + Docker

---

## 1. V1 功能范围

| 编号 | 功能模块 | 核心目标 |
|------|----------|----------|
| M1 | 项目工程搭建 | 前后端项目脚手架、开发规范、CI/CD 基础 |
| M2 | 用户认证与权限 | 登录/注册、JWT 鉴权、RBAC 角色权限 |
| M3 | 聊天控件 (Widget) | 可嵌入网站的实时聊天窗口 |
| M4 | 实时消息推送 | WebSocket 长连接、消息收发 |
| M5 | 客服后台收件箱 | 统一对话面板、对话列表 |
| M6 | 对话管理 | 消息收发、状态流转、对话分配 |
| M7 | 联系人管理 | 客户信息存储、对话关联 |
| M8 | 快捷回复 | 预设回复模板，一键插入 |
| M9 | REST API | 对外开放 API 接口 |

---

## 2. 详细任务分解

### M1 - 项目工程搭建

#### 2.1 后端工程

| 任务 | 说明 |
|------|------|
| Go 项目初始化 | Go Modules、项目目录结构（handler/service/repository/model） |
| HTTP 框架选型 | Gin / Fiber，路由分组、中间件机制 |
| 数据库 ORM | GORM 或 sqlx，连接池配置 |
| 配置管理 | Viper，支持 YAML/ENV 多环境配置 |
| 日志系统 | Zap / Zerolog，分级日志、请求追踪 ID |
| 错误码规范 | 统一错误码定义与响应格式 |
| Dockerfile | 后端多阶段构建 |
| 开发环境 | Docker Compose 编排 PostgreSQL + Redis + Kafka |

**目录结构建议**：

```
server/
├── cmd/              # 入口
├── config/           # 配置
├── internal/
│   ├── handler/      # HTTP 处理层
│   ├── service/      # 业务逻辑层
│   ├── repository/   # 数据访问层
│   ├── model/        # 数据模型
│   └── middleware/   # 中间件（认证、日志、CORS）
├── pkg/              # 公共库
├── migrations/       # 数据库迁移脚本
├── Dockerfile
└── go.mod
```

#### 2.2 前端工程

| 任务 | 说明 |
|------|------|
| Vue 3 + TS 项目初始化 | Vite + Vue 3 + TypeScript + Vue Router + Pinia |
| UI 组件库 | Element Plus / Ant Design Vue |
| 目录结构设计 | 页面、组件、Store、API、路由 |
| HTTP 客户端封装 | Axios 封装、请求拦截、Token 刷新 |
| 开发规范 | ESLint + Prettier + Husky |
| Dockerfile | 前端 Nginx 部署 |

**目录结构建议**：

```
web/
├── src/
│   ├── api/          # API 接口
│   ├── components/   # 公共组件
│   ├── composables/  # 组合式函数
│   ├── layouts/      # 布局组件
│   ├── pages/        # 页面
│   ├── router/       # 路由
│   ├── stores/       # Pinia Store
│   └── utils/        # 工具函数
├── Dockerfile
└── vite.config.ts
```

---

### M2 - 用户认证与权限

#### 2.3 后端

| 任务 | 说明 |
|------|------|
| 用户表设计 | id, email, name, password_hash, role, avatar, status, created_at, updated_at |
| 注册接口 | POST /api/v1/auth/register，邮箱+密码注册 |
| 登录接口 | POST /api/v1/auth/login，返回 JWT Token |
| 登出接口 | POST /api/v1/auth/logout |
| Token 刷新 | POST /api/v1/auth/refresh |
| JWT 中间件 | 解析 Token，注入用户信息到 Context |
| RBAC 权限中间件 | 角色枚举（admin/agent），接口级权限校验 |
| 修改密码 | PUT /api/v1/auth/password |

**用户表 (users)**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| email | VARCHAR(255) | 唯一，登录账号 |
| name | VARCHAR(100) | 显示名称 |
| password_hash | VARCHAR(255) | bcrypt 哈希 |
| role | VARCHAR(20) | admin / agent |
| avatar_url | VARCHAR(500) | 头像地址 |
| status | VARCHAR(20) | active / disabled |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

#### 2.4 前端

| 任务 | 说明 |
|------|------|
| 登录页面 | 邮箱+密码表单 |
| Token 管理 | 存储到 localStorage，请求拦截自动携带 |
| 路由守卫 | 未登录跳转登录页 |
| 权限指令/函数 | v-permission，控制按钮显示 |
| 401 处理 | Token 过期自动刷新或跳转登录 |
| 顶部导航栏 | 用户信息、修改密码、退出登录 |

---

### M3 - 聊天控件 (Live Chat Widget)

#### 2.5 Widget SDK

| 任务 | 说明 |
|------|------|
| Widget 项目初始化 | 独立的小型 Vue 3 应用，支持 script 标签嵌入 |
| 聊天窗口 UI | 右下角气泡按钮 + 展开聊天面板 |
| 消息列表 | 消息气泡（客户/客服样式区分）、自动滚动到底部 |
| 输入框 | 文本输入 + 回车发送 |
| WebSocket 连接 | 自动建立连接，断线重连 |
| 初始化参数 | app_id (可选), widget_token，自定义颜色 |
| 对话前表单 | 收集姓名、邮箱（可配置是否启用） |
| 新消息提示 | 未读消息红点，新消息桌面通知（可选） |
| 响应式适配 | 移动端全屏模式，PC 端小窗模式 |

**嵌入方式**：

```html
<script src="https://your-domain.com/widget.js" data-token="xxx" data-color="#1890ff"></script>
```

**Widget 状态管理**：

```
状态机：idle → pre_chat_form → chatting → ended
```

---

### M4 - 实时消息推送

#### 2.6 WebSocket 服务

| 任务 | 说明 |
|------|------|
| WebSocket 握手 | HTTP Upgrade，验证 JWT Token |
| 连接管理 | Conn Map 管理所有在线连接（用户 ID → Conn） |
| 心跳机制 | PING/PONG，30s 间隔，超时 90s 断开 |
| 断线重连 | 客户端指数退避重连 |
| 消息协议 | JSON 格式，包含 type、payload、timestamp |
| 消息确认 | ACK 机制，确保消息不丢失 |
| 离线消息 | 客户不在线时消息落库，上线后拉取 |

**消息协议设计**：

```json
{
  "id": "msg-uuid",
  "type": "message | typing | read_receipt | status_change | system",
  "conversation_id": "conv-uuid",
  "sender_id": "user-uuid",
  "sender_type": "agent | customer",
  "content": {
    "type": "text | image | file",
    "body": "消息内容"
  },
  "timestamp": "2025-01-01T00:00:00Z"
}
```

#### 2.7 消息存储

| 任务 | 说明 |
|------|------|
| 消息表设计 | 存储所有对话消息 |
| 消息查询 | 按 conversation_id 分页拉取历史消息 |

**消息表 (messages)**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| conversation_id | UUID | 所属对话 |
| sender_id | UUID | 发送者 |
| sender_type | VARCHAR(10) | agent / customer |
| content_type | VARCHAR(20) | text / image / file |
| content | TEXT | 消息体 |
| status | VARCHAR(20) | sent / delivered / read |
| created_at | TIMESTAMP | |

---

### M5 - 客服后台收件箱

#### 2.8 后端 API

| 接口 | 说明 |
|------|------|
| GET /api/v1/conversations | 对话列表（分页、筛选） |
| GET /api/v1/conversations/:id | 对话详情（含消息列表） |
| POST /api/v1/conversations | 创建对话（API / Widget 调用） |
| PATCH /api/v1/conversations/:id/assign | 分配对话 |
| PATCH /api/v1/conversations/:id/status | 变更对话状态 |

**对话表 (conversations)**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| contact_id | UUID | 关联联系人 |
| assigned_agent_id | UUID | 分配客服 |
| status | VARCHAR(20) | pending / open / resolved / closed |
| subject | VARCHAR(255) | 对话主题 |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |
| resolved_at | TIMESTAMP | |

#### 2.9 前端页面

| 任务 | 说明 |
|------|------|
| 整体布局 | 左侧导航 + 对话列表 + 右侧对话详情（三栏） |
| 对话列表 | 头像、客户名、最后一条消息预览、时间、状态标签、未读标记 |
| 状态筛选 | 全部 / 待处理 / 处理中 / 已解决 |
| 搜索 | 对话内容、客户名搜索 |
| 未读标记 | 未读对话高亮 + 未读数角标 |
| 实时更新 | 新对话 WebSocket 推送，列表自动更新排序 |
| 空状态 | 各状态列表空态占位图 |

---

### M6 - 对话管理

#### 2.10 对话详情页

| 任务 | 说明 |
|------|------|
| 消息区域 | 消息时间线，客服/客户消息样式区分 |
| 消息气泡 | 文本、图片（点击放大）、文件（下载链接） |
| 输入区域 | 文本输入框 + 发送按钮，Ctrl+Enter 发送 |
| 快捷回复入口 | 快捷回复选择器，点击插入输入框 |
| 对话信息侧栏 | 联系人信息、对话状态、创建时间 |
| 操作区 | 分配客服、变更状态（待处理→处理中→已解决）、添加标签 |

#### 2.11 状态流转

```
客户发起 → [待处理 Pending]
           ↓ 客服领取/分配
         [处理中 Open]
           ↓ 问题解决
         [已解决 Resolved]
           ↓ 客户重新回复
         [处理中 Open]  ← 自动重新打开
           ↓ 手动关闭(可选)
         [已关闭 Closed]
```

#### 2.12 对话分配

| 任务 | 说明 |
|------|------|
| 手动领取 | 客服从待处理列表点击"我来处理" |
| 手动分配 | 管理员将对话分配给指定客服 |
| 轮询分配 | 新对话按顺序分配给在线客服 |

---

### M7 - 联系人管理

#### 2.13 后端

| 任务 | 说明 |
|------|------|
| 联系人表设计 | 存储客户信息 |
| CRUD 接口 | 创建、查询、更新、删除联系人 |
| 关联对话查询 | 查询某联系人的所有历史对话 |
| 去重处理 | 同邮箱/手机号不重复创建 |

**联系人表 (contacts)**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | VARCHAR(100) | 姓名 |
| email | VARCHAR(255) | 邮箱 |
| phone | VARCHAR(30) | 手机号 |
| avatar_url | VARCHAR(500) | 头像 |
| source | VARCHAR(20) | 来源（widget / api） |
| notes | TEXT | 备注 |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

#### 2.14 前端

| 任务 | 说明 |
|------|------|
| 联系人列表页 | 分页表格、搜索、筛选 |
| 联系人详情 | 基本信息、历史对话列表 |
| 侧栏快捷查看 | 对话详情页侧栏查看联系人信息 |

---

### M8 - 快捷回复 (Canned Responses)

#### 2.15 后端

| 任务 | 说明 |
|------|------|
| 快捷回复表设计 | 存储回复模板 |
| CRUD 接口 | 个人模板和公共模板的增删改查 |

**快捷回复表 (canned_responses)**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| title | VARCHAR(200) | 标题（用于搜索） |
| content | TEXT | 回复内容 |
| short_code | VARCHAR(50) | 快捷指令，如 /hello |
| scope | VARCHAR(10) | personal / global |
| agent_id | UUID | 创建者 |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

#### 2.16 前端

| 任务 | 说明 |
|------|------|
| 快捷回复管理页 | 列表 + 新建/编辑弹窗，仅管理员可创建公共模板 |
| 对话快捷插入 | 输入框左侧按钮弹出选择器，支持搜索 |
| 快捷指令 | 输入 `/` 自动触发快捷回复搜索 |

---

### M9 - REST API

#### 2.17 API 设计规范

| 规范项 | 说明 |
|------|------|
| 版本管理 | URL 路径版本 /api/v1/ |
| 统一响应格式 | `{ code: 0, message: "ok", data: {} }` |
| 错误码规范 | 统一错误码枚举定义 |
| 分页规范 | page / page_size 参数，返回 total / list |
| 鉴权方式 | Bearer Token (JWT) |
| API 文档 | Swagger / OpenAPI 自动生成（基于代码注释） |

**统一响应结构**：

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "request_id": "trace-id"
}
```

#### 2.18 V1 API 清单

| 模块 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 认证 | POST | /api/v1/auth/register | 注册 |
| 认证 | POST | /api/v1/auth/login | 登录 |
| 认证 | POST | /api/v1/auth/refresh | 刷新 Token |
| 认证 | PUT | /api/v1/auth/password | 修改密码 |
| 联系人 | GET | /api/v1/contacts | 联系人列表 |
| 联系人 | GET | /api/v1/contacts/:id | 联系人详情 |
| 联系人 | POST | /api/v1/contacts | 创建联系人 |
| 联系人 | PUT | /api/v1/contacts/:id | 更新联系人 |
| 对话 | GET | /api/v1/conversations | 对话列表 |
| 对话 | GET | /api/v1/conversations/:id | 对话详情 |
| 对话 | POST | /api/v1/conversations | 创建对话 |
| 对话 | PATCH | /api/v1/conversations/:id/assign | 分配对话 |
| 对话 | PATCH | /api/v1/conversations/:id/status | 变更状态 |
| 消息 | GET | /api/v1/conversations/:id/messages | 消息列表 |
| 消息 | POST | /api/v1/conversations/:id/messages | 发送消息 |
| 快捷回复 | GET | /api/v1/canned_responses | 列表 |
| 快捷回复 | POST | /api/v1/canned_responses | 创建 |
| 快捷回复 | PUT | /api/v1/canned_responses/:id | 更新 |
| 快捷回复 | DELETE | /api/v1/canned_responses/:id | 删除 |
| WebSocket | WS | /ws | 实时消息 |

---

## 3. 数据库 ER 概览

```
┌───────────┐       ┌────────────────┐       ┌───────────┐
│  contacts  │──1:N──│ conversations  │──1:N──│  messages  │
└───────────┘       └───────┬────────┘       └───────────┘
                            │ N:1
                     ┌──────┴──────┐
                     │    users     │
                     │  (agents)    │
                     └─────────────┘

┌──────────────────┐
│ canned_responses │  (scope: personal/global, agent_id)
└──────────────────┘
```

---

## 4. 里程碑 & 时间规划

| 里程碑 | 内容 | 建议工期 |
|--------|------|----------|
| **Sprint 1** | M1 工程搭建 + M2 认证权限 | 1 周 |
| **Sprint 2** | M4 WebSocket + M5 收件箱后端 | 1.5 周 |
| **Sprint 3** | M5 收件箱前端 + M6 对话管理 | 1.5 周 |
| **Sprint 4** | M3 聊天控件 Widget | 1 周 |
| **Sprint 5** | M7 联系人管理 + M8 快捷回复 | 1 周 |
| **Sprint 6** | M9 REST API 完善 + 联调测试 | 1 周 |

**预计总工期**：6-7 周

---

## 5. 验收标准

| 标准 | 描述 |
|------|------|
| 完整流程 | 客户通过 Widget 发起对话 → 客服在后台收到消息 → 实时回复 → 解决后关闭 |
| 多会话并发 | 单个客服同时处理 10+ 个对话，消息实时无延迟（< 500ms） |
| 离线恢复 | 客服关闭页面重开后，自动重连 WebSocket，未读消息准确显示 |
| 权限隔离 | Agent 角色看不到管理功能入口，API 接口权限校验生效 |
| API 文档 | Swagger 文档完整，可直接调试 |

---

## 6. 风险与注意事项

| 风险 | 应对方案 |
|------|----------|
| WebSocket 高并发连接 | 使用 Go 协程管理连接，合理设置连接池上限 |
| 消息可靠性 | 发送失败重试机制，离线消息落库 |
| 前后端协议对齐 | 先约定 WebSocket 消息协议和 API 接口文档，再并行开发 |
| 数据库性能 | 消息表按时间分区或按月分表，建立合理索引 |
