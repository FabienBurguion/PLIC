ALTER TABLE matches
    ADD COLUMN creator_id TEXT NOT NULL DEFAULT 'dcdbe036-ee22-4f73-80be-b4bf6ae65539';

ALTER TABLE matches
    ADD CONSTRAINT fk_matches_creator
        FOREIGN KEY (creator_id)
            REFERENCES users(id)
            ON DELETE CASCADE;