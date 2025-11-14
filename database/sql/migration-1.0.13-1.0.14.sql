ALTER TABLE ranking
    ADD COLUMN sport sport NOT NULL DEFAULT 'basket';

ALTER TABLE ranking
    DROP CONSTRAINT ranking_user_court_unique,
    ADD CONSTRAINT ranking_user_court_sport_unique
        UNIQUE (user_id, court_id, sport);

CREATE INDEX IF NOT EXISTS idx_ranking_court_sport_elo
    ON ranking (court_id, sport, elo DESC);