CREATE TABLE IF NOT EXISTS dialogs (
  id BIGSERIAL PRIMARY KEY,
  a_id BIGINT NOT NULL,
  b_id BIGINT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- уникальность пары пользователей (без порядка)
CREATE UNIQUE INDEX IF NOT EXISTS uniq_dialog_pair
ON dialogs (LEAST(a_id,b_id), GREATEST(a_id,b_id));

CREATE TABLE IF NOT EXISTS messages (
  id BIGSERIAL PRIMARY KEY,
  dialog_id BIGINT NOT NULL,
  sender_id BIGINT NOT NULL,
  text TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  read_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_dialogs_user ON dialogs(a_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_dialogs_user_b ON dialogs(b_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_dialog ON messages(dialog_id, id);

