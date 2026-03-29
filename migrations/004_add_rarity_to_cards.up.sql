ALTER TABLE cards
    ADD COLUMN IF NOT EXISTS rarity TEXT NOT NULL DEFAULT 'common'
        CHECK (rarity IN ('common', 'uncommon', 'rare', 'ultra_rare'));
