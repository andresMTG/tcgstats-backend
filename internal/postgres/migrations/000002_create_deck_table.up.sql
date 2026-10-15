CREATE TYPE game_name AS ENUM (
    'magic the gathering',
    'pokemon',
    'flesh and blood',
    'yugioh',
    'digimon',
    'one piece',
    'lorcana'
);

CREATE TABLE IF NOT EXISTS decks (
    uuid UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    game game_name NOT NULL,
    decklist VARCHAR(255) NOT NULL, 
    user_uuid UUID REFERENCES users(uuid) ON DELETE CASCADE
);