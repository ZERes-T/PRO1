-- Колонки-счётчики на videos
ALTER TABLE videos
  ADD COLUMN IF NOT EXISTS comments_count  INT     NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS favorites_count INT     NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS views_count     BIGINT  NOT NULL DEFAULT 0;

-- Функция: апдейт updated_at диалога при новом сообщении
CREATE OR REPLACE FUNCTION bump_dialog_updated_at()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  UPDATE dialogs SET updated_at = NEW.created_at WHERE id = NEW.dialog_id;
  RETURN NEW;
END$$;

DROP TRIGGER IF EXISTS trg_messages_bump_dialog ON messages;
CREATE TRIGGER trg_messages_bump_dialog
AFTER INSERT ON messages
FOR EACH ROW EXECUTE FUNCTION bump_dialog_updated_at();

-- Функции: инк/дек счётчиков комментов и избранного
CREATE OR REPLACE FUNCTION videos_comments_counter()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE videos SET comments_count = comments_count + 1 WHERE id = NEW.video_id;
    RETURN NEW;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE videos SET comments_count = GREATEST(0, comments_count - 1) WHERE id = OLD.video_id;
    RETURN OLD;
  END IF;
  RETURN NULL;
END$$;

CREATE OR REPLACE FUNCTION videos_favorites_counter()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE videos SET favorites_count = favorites_count + 1 WHERE id = NEW.video_id;
    RETURN NEW;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE videos SET favorites_count = GREATEST(0, favorites_count - 1) WHERE id = OLD.video_id;
    RETURN OLD;
  END IF;
  RETURN NULL;
END$$;

DROP TRIGGER IF EXISTS trg_comments_counter_ins ON comments;
DROP TRIGGER IF EXISTS trg_comments_counter_del ON comments;
CREATE TRIGGER trg_comments_counter_ins
AFTER INSERT ON comments
FOR EACH ROW EXECUTE FUNCTION videos_comments_counter();
CREATE TRIGGER trg_comments_counter_del
AFTER DELETE ON comments
FOR EACH ROW EXECUTE FUNCTION videos_comments_counter();

DROP TRIGGER IF EXISTS trg_fav_counter_ins ON favorites;
DROP TRIGGER IF EXISTS trg_fav_counter_del ON favorites;
CREATE TRIGGER trg_fav_counter_ins
AFTER INSERT ON favorites
FOR EACH ROW EXECUTE FUNCTION videos_favorites_counter();
CREATE TRIGGER trg_fav_counter_del
AFTER DELETE ON favorites
FOR EACH ROW EXECUTE FUNCTION videos_favorites_counter();

-- Функция для инкремента просмотров из приложения (если захочешь вызывать SQL-ом)
CREATE OR REPLACE FUNCTION increment_video_views(vid BIGINT)
RETURNS VOID LANGUAGE plpgsql AS $$
BEGIN
  UPDATE videos SET views_count = views_count + 1 WHERE id = vid;
END$$;

