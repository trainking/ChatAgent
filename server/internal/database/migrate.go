package database

import "github.com/jmoiron/sqlx"

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'agent',
    avatar_url VARCHAR(500) DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    must_change_password BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`

const migration1 = `
ALTER TABLE users ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT true;
`

const permissionsSchema = `
CREATE TABLE IF NOT EXISTS permissions (
    code VARCHAR(100) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role VARCHAR(20) NOT NULL,
    permission_code VARCHAR(100) NOT NULL REFERENCES permissions(code) ON DELETE CASCADE,
    PRIMARY KEY (role, permission_code)
);
`

const seedPermissions = `
INSERT INTO permissions (code, name, description) VALUES
    ('inbox.view', '查看收件箱', 'View inbox and conversations'),
    ('inbox.reply', '回复消息', 'Reply to conversations'),
    ('inbox.assign', '分配对话', 'Assign conversations to agents'),
    ('users.manage', '账号管理', 'Create, edit and delete user accounts'),
    ('reports.view', '查看报表', 'View reports and analytics'),
    ('settings.view', '查看设置', 'View system settings'),
    ('roles.manage', '角色管理', 'Manage role permissions')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (role, permission_code)
SELECT 'admin', code FROM permissions WHERE code IN
    ('inbox.view', 'inbox.reply', 'inbox.assign', 'users.manage', 'reports.view')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role, permission_code)
SELECT 'agent', code FROM permissions WHERE code IN
    ('inbox.view', 'inbox.reply')
ON CONFLICT DO NOTHING;
`

const contactsSchema = `
CREATE TABLE IF NOT EXISTS contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL DEFAULT '',
    email VARCHAR(255) DEFAULT '',
    phone VARCHAR(32) DEFAULT '',
    avatar_url VARCHAR(500) DEFAULT '',
    contact_type VARCHAR(20) NOT NULL DEFAULT 'visitor',
    blocked BOOLEAN NOT NULL DEFAULT false,
    browser_fingerprints JSONB NOT NULL DEFAULT '[]',
    country VARCHAR(100) DEFAULT '',
    city VARCHAR(100) DEFAULT '',
    browser VARCHAR(100) DEFAULT '',
    os VARCHAR(100) DEFAULT '',
    custom_attrs JSONB NOT NULL DEFAULT '{}',
    last_activity_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`

const contactInboxesSchema = `
CREATE TABLE IF NOT EXISTS contact_inboxes (
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    inbox_id UUID NOT NULL REFERENCES inboxes(id) ON DELETE CASCADE,
    source_id VARCHAR(200) NOT NULL DEFAULT '',
    pubsub_token VARCHAR(255) NOT NULL DEFAULT '',
    hmac_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (contact_id, inbox_id)
);
`

const conversationsSchema = `
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    display_id INT NOT NULL DEFAULT 0,
    inbox_id UUID NOT NULL REFERENCES inboxes(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    priority VARCHAR(10) NOT NULL DEFAULT 'medium',
    subject VARCHAR(255) NOT NULL DEFAULT '',
    unread_count INT NOT NULL DEFAULT 0,
    last_message_at TIMESTAMP NOT NULL DEFAULT NOW(),
    waiting_since TIMESTAMP,
    first_reply_at TIMESTAMP,
    resolved_at TIMESTAMP,
    snoozed_until TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`

const messagesSchema = `
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_type VARCHAR(20) NOT NULL DEFAULT 'contact',
    sender_id UUID NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    message_type VARCHAR(20) NOT NULL DEFAULT 'incoming',
    content_type VARCHAR(50) NOT NULL DEFAULT 'text/html',
    file_url VARCHAR(1000) DEFAULT '',
    file_name VARCHAR(255) DEFAULT '',
    file_size BIGINT DEFAULT 0,
    status VARCHAR(10) NOT NULL DEFAULT 'sent',
    private BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`

const cannedResponsesSchema = `
CREATE TABLE IF NOT EXISTS canned_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inbox_id UUID REFERENCES inboxes(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`

const v2Indexes = `
CREATE INDEX IF NOT EXISTS idx_conversations_inbox_status_time ON conversations(inbox_id, status, last_message_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_conversations_inbox_display_id ON conversations(inbox_id, display_id);
CREATE INDEX IF NOT EXISTS idx_conversations_assignee ON conversations(assignee_id, status);
CREATE INDEX IF NOT EXISTS idx_conversations_contact ON conversations(contact_id);
CREATE INDEX IF NOT EXISTS idx_messages_conversation_created ON messages(conversation_id, created_at);
CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id, id);
CREATE INDEX IF NOT EXISTS idx_contacts_email ON contacts(email) WHERE email != '';
CREATE INDEX IF NOT EXISTS idx_contacts_phone ON contacts(phone) WHERE phone != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_contact_inboxes_source ON contact_inboxes(inbox_id, source_id);
CREATE INDEX IF NOT EXISTS idx_contact_inboxes_token ON contact_inboxes(pubsub_token);
`

func Migrate(db *sqlx.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	if _, err := db.Exec(migration1); err != nil {
		return err
	}
	if _, err := db.Exec(migration2); err != nil {
		return err
	}
	if _, err := db.Exec(migration3); err != nil {
		return err
	}
	if _, err := db.Exec(permissionsSchema); err != nil {
		return err
	}
	if _, err := db.Exec(seedPermissions); err != nil {
		return err
	}
	if _, err := db.Exec(totpSchema); err != nil {
		return err
	}
	if _, err := db.Exec(activityLogSchema); err != nil {
		return err
	}
	if _, err := db.Exec(inboxSchema); err != nil {
		return err
	}
	if _, err := db.Exec(contactsSchema); err != nil {
		return err
	}
	if _, err := db.Exec(contactInboxesSchema); err != nil {
		return err
	}
	if _, err := db.Exec(conversationsSchema); err != nil {
		return err
	}
	if _, err := db.Exec(messagesSchema); err != nil {
		return err
	}
	if _, err := db.Exec(cannedResponsesSchema); err != nil {
		return err
	}
	if _, err := db.Exec(v2Indexes); err != nil {
		return err
	}
	return nil
}

const migration2 = `
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP;
`

const migration3 = `
ALTER TABLE users ADD COLUMN IF NOT EXISTS online_status VARCHAR(20) NOT NULL DEFAULT 'offline';
`

const activityLogSchema = `
CREATE TABLE IF NOT EXISTS activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event VARCHAR(100) NOT NULL,
    module VARCHAR(100) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at);
`

const totpSchema = `
CREATE TABLE IF NOT EXISTS system_config (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_totp (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    secret VARCHAR(100) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO system_config (key, value) VALUES ('2fa_enabled', 'false') ON CONFLICT (key) DO NOTHING;
INSERT INTO system_config (key, value) VALUES ('2fa_issuer', 'ChatAgent') ON CONFLICT (key) DO NOTHING;
`

const inboxSchema = `
CREATE TABLE IF NOT EXISTS inboxes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    welcome_title VARCHAR(32) DEFAULT '',
    welcome_message TEXT DEFAULT '',
    inbox_type VARCHAR(20) NOT NULL CHECK (inbox_type IN ('website', 'api')),
    status VARCHAR(20) NOT NULL DEFAULT 'enabled',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inbox_collaborators (
    inbox_id UUID NOT NULL REFERENCES inboxes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (inbox_id, user_id)
);
`
