CREATE TABLE IF NOT EXISTS match_score_vote (
    match_id TEXT REFERENCES matches(id) ON DELETE CASCADE,
    user_id  TEXT REFERENCES users(id)   ON DELETE CASCADE,
    team     INTEGER NOT NULL CHECK (team IN (1,2)),
    score1   INTEGER NOT NULL,
    score2   INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (match_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_score_vote_match_team_score
    ON match_score_vote (match_id, team, score1, score2);

CREATE OR REPLACE FUNCTION try_finalize_match() RETURNS trigger AS $$
DECLARE
    other_team INT;
    agree_exists BOOLEAN;
BEGIN
    IF NEW.team = 1 THEN other_team := 2; ELSE other_team := 1; END IF;

    SELECT EXISTS (
        SELECT 1
        FROM match_score_vote v
        WHERE v.match_id = NEW.match_id
          AND v.team = other_team
          AND v.score1 = NEW.score1
          AND v.score2 = NEW.score2
    ) INTO agree_exists;

    IF agree_exists THEN
        UPDATE matches
        SET score1 = NEW.score1,
            score2 = NEW.score2,
            current_state = 'Termine',
            updated_at = NOW()
        WHERE id = NEW.match_id
          AND current_state = 'Manque Score';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_try_finalize_match ON match_score_vote;
CREATE TRIGGER trg_try_finalize_match
    AFTER INSERT OR UPDATE ON match_score_vote
    FOR EACH ROW EXECUTE FUNCTION try_finalize_match();