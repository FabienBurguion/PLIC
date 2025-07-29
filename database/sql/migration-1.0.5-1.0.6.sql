ALTER TABLE terrain
    RENAME TO courts;

ALTER TABLE classement RENAME COLUMN terrain_id TO court_id;

ALTER TABLE classement RENAME TO ranking;


ALTER TABLE matches RENAME COLUMN lieu TO place;
ALTER TABLE matches RENAME COLUMN nbre_participant TO participant_nber;
ALTER TABLE matches RENAME COLUMN etat TO current_state;
ALTER TABLE matches RENAME COLUMN terrain_id TO court_id;