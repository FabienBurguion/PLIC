CREATE TYPE dayOfWeek AS ENUM (
  'Lundi', 'Mardi', 'Mercredi', 'Jeudi', 'Vendredi', 'Samedi', 'Dimanche'
);

CREATE TYPE typeOfSport AS ENUM (
  'basket', 'tennis de table', 'foot'
);

CREATE TYPE friendStatus AS ENUM (
  'en attente', 'accepté', 'refusé'
);

CREATE TABLE users (
                       id text PRIMARY KEY,
                       name varchar(20) NOT NULL,
                       surname varchar(30) NOT NULL,
                       email varchar(30) UNIQUE NOT NULL,
                       hPassword text NOT NULL,
                       birthdate timestamp NOT NULL CHECK (birthdate < NOW()),
                       description text,
                       city varchar(45),
                       photoUrl text DEFAULT 'nophoto.jpg',
                       CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._-]+@[A-Za-z0-9-]+\.[A-Za-z]{2,}$')
);

CREATE TABLE courts (
                        id text PRIMARY KEY,
                        address text NOT NULL,
                        name varchar(20) NOT NULL,
                        website varchar(50),
                        longitude float NOT NULL,
                        latitude float NOT NULL,
                        CONSTRAINT valid_longitude CHECK (longitude BETWEEN -180 AND 180),
                        CONSTRAINT valid_latitude CHECK (latitude BETWEEN -90 AND 90)
);

CREATE TABLE matches (
                         id text PRIMARY KEY,
                         dateOf timestamp NOT NULL,
                         court_id text REFERENCES courts(id) ON DELETE SET NULL,
                         score varchar(10) DEFAULT '0-0',
                         CONSTRAINT valid_score CHECK (score ~ '^[0-9]+-[0-9]+$')
    );

CREATE TABLE openingHours (
                              id text PRIMARY KEY,
                              court_id text REFERENCES courts(id) ON DELETE CASCADE,
                              dayOfWeek dayOfWeek NOT NULL,
                              openTime timestamp NOT NULL,
                              closeTime timestamp NOT NULL
);

CREATE TABLE sports (
                        type typeOfSport PRIMARY KEY
);

CREATE TABLE user_match (
                            user_id text REFERENCES users(id) ON DELETE CASCADE,
                            match_id text REFERENCES matches(id) ON DELETE CASCADE,
                            PRIMARY KEY (user_id, match_id)
);

CREATE TABLE court_sport (
                             court_id text REFERENCES courts(id) ON DELETE CASCADE,
                             sport_id typeOfSport REFERENCES sports(type) ON DELETE CASCADE,
                             PRIMARY KEY (court_id, sport_id),
                             UNIQUE (court_id, sport_id)
);

CREATE TABLE friendship (
                            user_id text REFERENCES users(id) ON DELETE CASCADE,
                            other_id text REFERENCES users(id) ON DELETE CASCADE,
                            status friendStatus NOT NULL DEFAULT 'en attente',
                            PRIMARY KEY (user_id, other_id)
);