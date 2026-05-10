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
	return nil
}

const migration2 = `
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP;
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
