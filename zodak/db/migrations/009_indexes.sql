-- уже были, но добавим полезные

-- FAVORITES
CREATE INDEX IF NOT EXISTS idx_fav_user  ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_fav_video ON favorites(video_id);

-- COMMENTS (by user)
CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id, created_at);

-- MESSAGES (по отправителю)
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id, id);

-- REQUESTS (по статусу)
CREATE INDEX IF NOT EXISTS idx_requests_status ON requests(status, created_at DESC);

-- QUOTES (по статусу)
CREATE INDEX IF NOT EXISTS idx_quotes_status ON quotes(status, created_at DESC);

