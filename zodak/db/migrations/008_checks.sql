-- ЗАЯВКИ: только нужные статусы
ALTER TABLE requests
  ADD CONSTRAINT chk_requests_status CHECK (status IN ('new','answered','archived'));

-- ВИДЕО: статус ограничим, длительность и тайтл
ALTER TABLE videos
  ADD CONSTRAINT chk_videos_status CHECK (status IN ('ready','pending','live','deleted')),
  ADD CONSTRAINT chk_videos_title  CHECK (char_length(title) > 0),
  ALTER COLUMN duration_sec SET DEFAULT 0,
  ALTER COLUMN duration_sec SET NOT NULL;

-- КОММЕНТЫ: не пустой текст
ALTER TABLE comments
  ADD CONSTRAINT chk_comments_text CHECK (char_length(text) > 0);

-- QUOTES (конфигуратор): валидные диапазоны + статусы + материалы/фасады
ALTER TABLE quotes
  ADD CONSTRAINT chk_quotes_dims CHECK (w_mm BETWEEN 500 AND 5000 AND h_mm BETWEEN 500 AND 3000 AND d_mm BETWEEN 300 AND 1200),
  ADD CONSTRAINT chk_quotes_doors CHECK (doors BETWEEN 1 AND 6),
  ADD CONSTRAINT chk_quotes_status CHECK (status IN ('draft','sent','accepted','archived')),
  ADD CONSTRAINT chk_quotes_material CHECK (material IN ('ldsp','mdf','plywood')),
  ADD CONSTRAINT chk_quotes_facade   CHECK (facade IN ('none','mirror','glass','mdf_paint'));

-- ТИКЕТЫ: статусы
ALTER TABLE support_tickets
  ADD CONSTRAINT chk_support_status CHECK (status IN ('open','pending','closed'));

-- Емейл уникален без учета регистра (дополнительно к текущему UNIQUE(email))
CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_idx ON users (lower(email));

