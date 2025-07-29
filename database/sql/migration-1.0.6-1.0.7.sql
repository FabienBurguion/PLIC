ALTER TABLE user_match
DROP CONSTRAINT IF EXISTS user_match_match_id_fkey;

ALTER TABLE user_match
    ADD CONSTRAINT user_match_match_id_fkey
        FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE;
