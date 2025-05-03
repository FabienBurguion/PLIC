CREATE TABLE IF NOT EXISTS users (
     id TEXT PRIMARY KEY,
     username TEXT NOT NULL,
     password TEXT NOT NULL,
     created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS terrain (
                                     id TEXT PRIMARY KEY,
                                     address TEXT NOT NULL,
                                     longitude DOUBLE PRECISION NOT NULL,
                                     latitude DOUBLE PRECISION NOT NULL
    );