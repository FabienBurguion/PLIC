CREATE TABLE IF NOT EXISTS users (
     id TEXT PRIMARY KEY,
     username TEXT UNIQUE NOT NULL,
     email TEXT UNIQUE NOT NULL,
     bio TEXT,
     current_field_id TEXT,
     password TEXT NOT NULL,
     created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS courts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
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
   place TEXT NOT NULL,
   date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
   participant_nber INTEGER NOT NULL DEFAULT 0,
   current_state etat_match NOT NULL DEFAULT 'Manque joueur',
   score1 INTEGER NOT NULL DEFAULT -1,
   score2 INTEGER NOT NULL DEFAULT -1,
   court_id TEXT REFERENCES courts(id),
   created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ranking (
   user_id TEXT REFERENCES users(id),
   court_id TEXT REFERENCES courts(id),
   elo INTEGER NOT NULL DEFAULT 200,
   UNIQUE (user_id, court_id),
   created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_match (
  user_id TEXT REFERENCES users(id),
  match_id TEXT REFERENCES matches(id) ON DELETE CASCADE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
