ALTER TABLE links
ADD CONSTRAINT unique_user_id_description UNIQUE (user_id, description);