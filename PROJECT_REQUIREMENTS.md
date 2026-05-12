# ChatAgent Project Requirements

## Product Goal

ChatAgent is a Go/Vue implementation of the core Chatwoot experience: a live-chat customer support platform with an embeddable website widget, an agent conversation workspace, contact management, inbox management, role-based administration, and real-time messaging.

The near-term goal is not to clone every Chatwoot feature. The first reliable milestone is a production-usable live chat loop:

Visitor widget -> widget auth -> contact/contact inbox -> conversation -> message persistence -> agent inbox -> agent reply -> real-time widget update.

## Current Architecture

| Area | Stack | Status |
| --- | --- | --- |
| Server | Go, Gin, sqlx, PostgreSQL, Redis, JWT, WebSocket | Active implementation |
| Web app | Vue 3, TypeScript, Vite, Element Plus, Pinia | Active implementation |
| Widget | Native TypeScript bundled by esbuild | Active implementation |
| Database | Inline idempotent SQL in `server/internal/database/migrate.go` | No migration versioning yet |
| Realtime | Gorilla WebSocket in `server/internal/websocket` | Single-node hub implemented |

## Implemented

- First-run root user initialization.
- JWT login/logout and forced password change flow.
- Optional TOTP 2FA.
- User management for admin roles.
- Basic role permissions screen for super admins.
- Inbox CRUD with collaborators and website/API inbox type.
- Contact list, create, edit, detail, merge, and conversation history APIs.
- Conversation list/detail, status, priority, assignee, unread count, and messages.
- File upload endpoint and attachment message support.
- Website widget auth, visitor contact creation, message send, and local widget UI.
- WebSocket hub for agent/widget subscriptions and message broadcasts.
- Basic Chinese/English i18n in the web app.

## Recently Stabilized

- Agent conversation filtering now treats `assignee_id=me` as the current user, not every assigned conversation.
- Non-admin agents are restricted to conversations in inboxes they created or collaborate on.
- WebSocket subscriptions are checked before joining channels:
  - agents may subscribe only to allowed inbox/conversation channels;
  - widgets may subscribe only to their own conversations.
- Attachment messages no longer require text content.
- Agent replies move conversations to `pending`; customer replies can reopen resolved conversations.
- Widget now maps `conversation_id` from the server response correctly and subscribes to the conversation channel after sending.
- Widget WebSocket connects back to the widget server origin, not the host page origin.
- Widget left-bubble and embedded modes re-render on store updates.
- Contact merge no longer uses invalid PostgreSQL syntax.
- Basic message HTML sanitization was added in the web conversation view.

## Known Gaps

- No automated tests cover business flows yet.
- No migration versioning; schema changes are startup-only DDL.
- No multi-account/tenant model.
- No team management, labels, custom views, assignment rules, automation, SLA, reports, CSAT, or help center.
- No channel adapters beyond website widget/API-shaped inboxes.
- WebSocket is single-node only; Redis Pub/Sub fanout is not implemented.
- Rich text sanitization is basic and should become a shared server/client policy.
- Permission checks are incomplete on some REST handlers, especially contact detail/update and conversation mutations.
- Widget has no message history sync on reload.
- `web` build has a large chunk warning that should be addressed with manual chunking later.

## Priority Roadmap

### P0 - Make The Live Chat Loop Trustworthy

- Add integration tests for widget auth, customer message, agent reply, WebSocket subscription, and contact merge.
- Complete REST permission checks for conversations, contacts, messages, and inbox access.
- Add widget message history fetch/recovery.
- Add server-side HTML sanitization or a strict allowed-rich-text format.
- Add conversation activity messages for assignment/status changes.

### P1 - Agent Workspace Completeness

- Improve conversation UI ergonomics: assignee picker, contact sidebar, message status, unread reset feedback.
- Add canned responses.
- Add internal notes with clearer visibility rules.
- Add typing indicators end to end.
- Add saved filters/custom views.

### P2 - Chatwoot-Like Operations

- Teams and assignment rules.
- Labels and customer segments.
- Reports and CSAT.
- Webhooks and public API keys.
- Redis Pub/Sub support for multi-server WebSocket deployment.

### P3 - Channels And Platform

- Email channel.
- WeChat/enterprise messaging adapters.
- Help center.
- AI reply assistance and conversation summarization.
- SSO/SAML.

## Development Rules

- Keep server API responses in the existing `{ code, message, data }` shape.
- Add schema changes to `server/internal/database/migrate.go` until a migration framework is introduced.
- Prefer focused fixes over broad rewrites.
- Run these before handing off substantial changes:
  - `cd server && go test ./...`
  - `cd web && npm run build`
  - `cd widget && npm run build`
