CREATE TYPE tournament_scope AS ENUM (
    'LOCAL',
    'QUALIFIER',
    'MAJOR',
    'PREMIER'
);

CREATE TABLE IF NOT EXISTS tournaments (
    uuid UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    game game_name NOT NULL,
    scope tournament_scope NOT NULL,
    is_official_circuit BOOLEAN DEFAULT false,
    user_uuid UUID REFERENCES users(uuid) ON DELETE CASCADE,
    deck_uuid UUID REFERENCES decks(uuid) ON DELETE CASCADE
);
