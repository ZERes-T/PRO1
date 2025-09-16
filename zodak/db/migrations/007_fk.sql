-- USERS
-- уже есть: users(id)

-- FOLLOWS
ALTER TABLE follows
  ADD CONSTRAINT fk_follows_user   FOREIGN KEY (user_id)   REFERENCES users(id) ON DELETE CASCADE,
  ADD CONSTRAINT fk_follows_target FOREIGN KEY (target_id) REFERENCES users(id) ON DELETE CASCADE;

-- VIDEOS
ALTER TABLE videos
  ADD CONSTRAINT fk_videos_author FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE;

-- COMMENTS
ALTER TABLE comments
  ADD CONSTRAINT fk_comments_video FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE,
  ADD CONSTRAINT fk_comments_user  FOREIGN KEY (user_id)  REFERENCES users(id) ON DELETE CASCADE;

-- FAVORITES
ALTER TABLE favorites
  ADD CONSTRAINT fk_fav_user  FOREIGN KEY (user_id)  REFERENCES users(id) ON DELETE CASCADE,
  ADD CONSTRAINT fk_fav_video FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE;

-- REQUESTS
ALTER TABLE requests
  ADD CONSTRAINT fk_requests_author FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE;

-- STREAM KEYS
ALTER TABLE stream_keys
  ADD CONSTRAINT fk_stream_keys_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- DIALOGS / MESSAGES
ALTER TABLE dialogs
  ADD CONSTRAINT fk_dialogs_a FOREIGN KEY (a_id) REFERENCES users(id) ON DELETE CASCADE,
  ADD CONSTRAINT fk_dialogs_b FOREIGN KEY (b_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE messages
  ADD CONSTRAINT fk_messages_dialog FOREIGN KEY (dialog_id) REFERENCES dialogs(id) ON DELETE CASCADE,
  ADD CONSTRAINT fk_messages_sender FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE;

-- QUOTES
ALTER TABLE quotes
  ADD CONSTRAINT fk_quotes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- SUPPORT
ALTER TABLE support_tickets
  ADD CONSTRAINT fk_support_tickets_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE support_messages
  ADD CONSTRAINT fk_support_messages_ticket FOREIGN KEY (ticket_id) REFERENCES support_tickets(id) ON DELETE CASCADE,
  ADD CONSTRAINT fk_support_messages_user   FOREIGN KEY (user_id)   REFERENCES users(id) ON DELETE CASCADE;

