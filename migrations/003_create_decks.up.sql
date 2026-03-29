CREATE TABLE IF NOT EXISTS decks (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_decks_user_id ON decks(user_id);

-- Each row is one card slot in a deck. quantity enforces how many copies are included.
-- Deck construction rules (min size, max copies) are enforced at the API layer until
-- those rules are finalised (see open items in architecture-spec.md).
CREATE TABLE IF NOT EXISTS deck_cards (
    deck_id  UUID NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    card_id  TEXT NOT NULL REFERENCES cards(id),
    quantity INT  NOT NULL DEFAULT 1 CHECK (quantity >= 1),
    PRIMARY KEY (deck_id, card_id)
);
