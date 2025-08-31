CREATE TABLE IF NOT EXISTS user_roles (
  user_id    BIGINT NOT NULL,
  role       TEXT   NOT NULL,   -- admin | moderator | support
  granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, role)
);

ALTER TABLE user_roles
  ADD CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE user_roles
  ADD CONSTRAINT chk_user_roles_role CHECK (role IN ('admin','moderator','support'));

CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role, user_id);

