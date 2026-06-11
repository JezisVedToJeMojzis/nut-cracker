-- Accounts. The id is a sequential BIGINT (the primary key, automatically
-- indexed) used as the short, shareable, orderable user id. username is the
-- editable display name.
CREATE TABLE IF NOT EXISTS users (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT,                -- nullable: Google-only users have no password
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- External login identities (Google now; email/password and others later).
CREATE TABLE IF NOT EXISTS user_identities (
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider         TEXT NOT NULL,    -- e.g. 'google'
    provider_user_id TEXT NOT NULL,    -- stable ID from the provider
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (provider, provider_user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_identities_user_id ON user_identities(user_id);

-- Reference list of world countries, keyed by ISO 3166-1 alpha-2 code.
CREATE TABLE IF NOT EXISTS countries (
    code TEXT PRIMARY KEY,             -- e.g. 'SK'
    name TEXT NOT NULL                 -- e.g. 'Slovakia'
);

-- Countries a user has "cracked", with how many people from that country.
CREATE TABLE IF NOT EXISTS user_countries (
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    country_code TEXT NOT NULL REFERENCES countries(code),
    cracks       INTEGER NOT NULL DEFAULT 1 CHECK (cracks >= 1),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, country_code)
);

-- Friend relationships (request + accept model).
CREATE TABLE IF NOT EXISTS friendships (
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, friend_id),
    CHECK (user_id <> friend_id)       -- can't befriend yourself
);

CREATE INDEX IF NOT EXISTS idx_friendships_friend_id ON friendships(friend_id);
