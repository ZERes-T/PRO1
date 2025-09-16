-- uploads: учёт загруженных файлов (в т.ч. сырого видео)
CREATE TABLE IF NOT EXISTS uploads (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  kind TEXT NOT NULL,           -- "video_raw" | "image" | "doc"
  url  TEXT NOT NULL,           -- s3://... или https://cdn...
  size_bytes BIGINT NOT NULL DEFAULT 0,
  mime TEXT NOT NULL DEFAULT '',
  meta JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE uploads
  ADD CONSTRAINT fk_uploads_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_uploads_user ON uploads(user_id, created_at DESC);

-- варианты видео (HLS)
CREATE TABLE IF NOT EXISTS video_variants (
  id BIGSERIAL PRIMARY KEY,
  video_id BIGINT NOT NULL,
  quality  TEXT NOT NULL,       -- "240p" | "480p" | "720p" | "1080p"
  playlist_url TEXT NOT NULL,   -- https://cdn.../master_720.m3u8
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE video_variants
  ADD CONSTRAINT fk_variants_video FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_video_variant ON video_variants(video_id, quality);

