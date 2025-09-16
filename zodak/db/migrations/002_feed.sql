-- Индексы для подписок
CREATE INDEX IF NOT EXISTS idx_follows_user ON follows(user_id);
CREATE INDEX IF NOT EXISTS idx_follows_target ON follows(target_id);

-- Автор сам на себя (чтобы свои видео попадали в ленту)
INSERT INTO follows(user_id, target_id)
SELECT 1, 1
WHERE NOT EXISTS (SELECT 1 FROM follows WHERE user_id=1 AND target_id=1);

-- Демо-видео для автора #1
INSERT INTO videos(author_id, title, descr, poster_url)
VALUES
(1, 'Демо-видео 1', 'Первое тестовое видео', ''),
(1, 'Демо-видео 2', 'Второе тестовое видео', ''),
(1, 'Демо-видео 3', 'Третье тестовое видео', ''),
(1, 'Демо-видео 4', 'Четвертое тестовое видео', ''),
(1, 'Демо-видео 5', 'Пятое тестовое видео', ''),
(1, 'Демо-видео 6', 'Шестое тестовое видео', ''),
(1, 'Демо-видео 7', 'Седьмое тестовое видео', '')
ON CONFLICT DO NOTHING;

