CREATE TABLE IF NOT EXISTS cards (
    id            TEXT        PRIMARY KEY,
    name          TEXT        NOT NULL,
    card_type     TEXT        NOT NULL CHECK (card_type IN ('conqueror', 'spell', 'constant', 'structure')),
    ap_cost       INT,                          -- NULL for structures (not played from hand)
    atk           INT,                          -- conquerors only
    def           INT,                          -- conquerors only
    hp            INT,                          -- conquerors and structures
    spd           INT,                          -- conquerors only
    rng           INT,                          -- conquerors only
    keywords      TEXT[]      NOT NULL DEFAULT '{}',
    effect_id     TEXT,
    effect_params JSONB,
    tags          TEXT[]      NOT NULL DEFAULT '{}',  -- e.g. 'immediate' for spells
    flavor_text   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
