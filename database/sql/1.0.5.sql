CREATE TABLE IF NOT EXISTS users (
     id TEXT PRIMARY KEY,
     username TEXT UNIQUE NOT NULL,
     email TEXT UNIQUE NOT NULL,
     bio TEXT,
     password TEXT NOT NULL,
     created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS terrain (
   id TEXT PRIMARY KEY,
   address TEXT NOT NULL,
   longitude DOUBLE PRECISION NOT NULL,
   latitude DOUBLE PRECISION NOT NULL,
   created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TYPE sport AS ENUM(
    'basket',
    'foot'
    );

CREATE TYPE etat_match AS ENUM(
    'Termine', -- match termine et score valide
    'Manque Score', -- score a valide mais match terminé
    'En cours', -- en train de faire le match
    'Valide', -- ts les participants on rejoint masi pas encore la date
    'Manque joueur' -- ts les participants n'ont pas encore rejoint
    );

CREATE TABLE IF NOT EXISTS matches (
    id TEXT PRIMARY KEY,
    sport sport NOT NULL DEFAULT 'basket',
    lieu TEXT NOT NULL,
    date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    nbre_participant INTEGER NOT NULL DEFAULT 0,
    etat etat_match NOT NULL DEFAULT 'Manque joueur',
    score1 INTEGER NOT NULL DEFAULT -1,
    score2 INTEGER NOT NULL DEFAULT -1,
    terrain_id TEXT REFERENCES terrain(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS classement (
    user_id TEXT REFERENCES users(id),
    terrain_id TEXT REFERENCES terrain(id),
    elo INTEGER NOT NULL DEFAULT 200,
    UNIQUE (user_id, terrain_id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS user_match (
    user_id TEXT REFERENCES users(id),
    match_id TEXT REFERENCES matches(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
