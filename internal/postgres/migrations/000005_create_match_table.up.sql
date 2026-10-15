CREATE TYPE result_name AS ENUM (
    'win',
    'draw',
    'lose'
);

CREATE TYPE match_phase AS ENUM (
    'SWISS',
    'TOP_8',
    'TOP_4',
    'FINALS'
);

CREATE TABLE IF NOT EXISTS matches (
    uuid UUID PRIMARY KEY,
    result result_name NOT NULL,
    round_number INT NOT NULL,
    phase match_phase NOT NULL,
    opposing_deck_archetype VARCHAR(255) NOT NULL,
    tournament_uuid UUID REFERENCES tournaments(uuid) ON DELETE CASCADE,
    opponent_uuid UUID REFERENCES opponents(uuid) ON DELETE CASCADE
);