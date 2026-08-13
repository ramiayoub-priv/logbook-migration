-- Logbook schema. See app/docs/data-model.md for the reasoning behind it and
-- app/docs/security.md for the auth tables.
--
-- Two rules are enforced here rather than only in Go, because a CHECK survives
-- a careless migration and a code comment does not:
--
--   * durations are whole non-negative minutes, never H:MM strings (rule 0.1
--     of data-model.md);
--   * there is no cumulative column anywhere (CLAUDE.md rule 0.5). Running
--     totals are derived from `seq` at query time.
--
-- Every statement is idempotent so that opening an existing database is a
-- no-op rather than a migration.

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_version (
    version    INTEGER NOT NULL,
    applied_at TEXT    NOT NULL
);

-- The seed list behind the new-flight form.
--
-- It was ONLY derived, rebuilt from the flights on every import -- which meant
-- the only aeroplanes that could exist were the ones already flown, and a first
-- flight in a new aeroplane was unenterable. Since 2026-08-02 it also holds rows
-- written by hand, marked user_added, and they survive an import.
--
-- IT IS A SEED LIST, NOT THE RECORD. Every flight carries its own registration,
-- type and class as written on paper, so nothing here can move a total. Editing
-- an aeroplane is therefore safe in a way that editing a flight is not.
CREATE TABLE IF NOT EXISTS aircraft (
    id            INTEGER PRIMARY KEY,
    registration  TEXT    NOT NULL UNIQUE,
    type          TEXT    NOT NULL,
    default_class TEXT    NOT NULL
                  CHECK (default_class IN ('SEP_LAND','SEP_SEA','MEP_LAND','MEP_SEA','TMG')),
    -- A hint for the form only. It never constrains what a flight may record.
    ifr_capable   INTEGER NOT NULL DEFAULT 0 CHECK (ifr_capable IN (0,1)),
    -- VESTIGIAL. The owner ruled on 2026-08-02 that there is no retired/active
    -- concept: an aeroplane you flew once in 2009 is not "retired", and the long
    -- list is solved by a filterable dropdown ordered by what was flown most
    -- recently. Nothing reads this column. It is left in place rather than
    -- migrated away, because dropping a column from a live legal-record schema
    -- to delete a feature is a worse trade than an unused column.
    active        INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0,1)),
    notes         TEXT    NOT NULL DEFAULT ''
    -- user_added is added by migrate() in store.go: SQLite has no
    -- ADD COLUMN IF NOT EXISTS, and every statement in this file must stay
    -- idempotent so that opening a live database is a no-op.
);

-- Names that may be written into a flight's PIC field but are not on any flight
-- yet. The roster the form offers is this table UNION the distinct pic_name
-- values already in the record -- so, like the aircraft list, it is mostly
-- derived and this table only holds what has not been flown with yet.
--
-- WHY IT EXISTS (owner ask, 2026-08-03): "I could have a typo when I write
-- `self`, it could be `sself` or `SELF` or `seeelf` and I need it to be
-- consistent (like the aircraft regs)." A free-text field is one keystroke away
-- from splitting 1143 flights across two spellings of the same word.
--
-- COLLATE NOCASE is what stops `SELF` joining `self`. It is ASCII-only, so it
-- would not catch a case variant of `Sinervä`'s ä; AddPilot's own check has the
-- same limit and says so. The constraint is here as well as in Go because a
-- UNIQUE index cannot be raced and a SELECT-then-INSERT can.
--
-- IT IS A SEED LIST, NOT THE RECORD, exactly like `aircraft`. Every flight
-- carries its own pic_name as written on paper; nothing here can change one.
CREATE TABLE IF NOT EXISTS pilots (
    id    INTEGER PRIMARY KEY,
    name  TEXT NOT NULL UNIQUE COLLATE NOCASE,
    notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS flights (
    id          INTEGER PRIMARY KEY,

    -- Explicit book order. Many flights share a date and 18 rows across the
    -- three books are genuinely out of date order on paper, so ordering must
    -- never fall back on flight_date. This is the key every cumulative
    -- computation walks.
    seq         INTEGER NOT NULL UNIQUE,
    flight_date TEXT    NOT NULL,

    -- Nullable on purpose: a flight is never lost because its aircraft row is
    -- missing. ON DELETE SET NULL for the same reason.
    aircraft_id   INTEGER REFERENCES aircraft(id) ON DELETE SET NULL,
    -- Denormalized as written on paper. If the aircraft table later disagrees
    -- that is a discrepancy to surface, not to auto-correct (rule 0.2).
    aircraft_reg  TEXT NOT NULL,
    aircraft_type TEXT NOT NULL,

    -- Per-flight, seeded from aircraft.default_class, user-overridable. This
    -- is what the sea/land statistics split on.
    class     TEXT NOT NULL
              CHECK (class IN ('SEP_LAND','SEP_SEA','MEP_LAND','MEP_SEA','TMG')),
    dep_place TEXT NOT NULL DEFAULT '',
    arr_place TEXT NOT NULL DEFAULT '',

    -- Canonical RFC3339 UTC, or NULL for a blank cell. RFC3339 sorts
    -- lexicographically, which is what range queries rely on.
    off_block_utc TEXT,
    on_block_utc  TEXT,
    -- Exactly as written on paper: "15:30" or "07:56Z". Never rewritten; this
    -- is what makes a conversion auditable (rule 0.4).
    off_block_raw TEXT NOT NULL DEFAULT '',
    on_block_raw  TEXT NOT NULL DEFAULT '',
    time_origin   TEXT NOT NULL
                  CHECK (time_origin IN ('utc_as_written','converted_from_local','unknown','none')),
    takeoff_utc   TEXT,
    landing_utc   TEXT,

    block_minutes      INTEGER NOT NULL DEFAULT 0 CHECK (block_minutes      >= 0),
    -- The flown time the book totals on. Equal to block_minutes on all but one
    -- row; never adjust one to match the other.
    total_minutes      INTEGER NOT NULL DEFAULT 0 CHECK (total_minutes      >= 0),
    night_minutes      INTEGER NOT NULL DEFAULT 0 CHECK (night_minutes      >= 0),
    instrument_minutes INTEGER NOT NULL DEFAULT 0 CHECK (instrument_minutes >= 0),
    pic_minutes        INTEGER NOT NULL DEFAULT 0 CHECK (pic_minutes        >= 0),
    dual_minutes       INTEGER NOT NULL DEFAULT 0 CHECK (dual_minutes       >= 0),
    instructor_minutes INTEGER NOT NULL DEFAULT 0 CHECK (instructor_minutes >= 0),
    -- Zero everywhere so far; the EASA book has the columns and the PDF must
    -- render them.
    copilot_minutes    INTEGER NOT NULL DEFAULT 0 CHECK (copilot_minutes    >= 0),
    multipilot_minutes INTEGER NOT NULL DEFAULT 0 CHECK (multipilot_minutes >= 0),

    pic_name          TEXT    NOT NULL DEFAULT '',
    landings_day      INTEGER NOT NULL DEFAULT 0 CHECK (landings_day   >= 0),
    landings_night    INTEGER NOT NULL DEFAULT 0 CHECK (landings_night >= 0),
    -- 0 = the day/night split was inferred, not read from paper. Drives the
    -- review list.
    landings_verified INTEGER NOT NULL DEFAULT 0 CHECK (landings_verified IN (0,1)),
    remarks           TEXT    NOT NULL DEFAULT '',

    -- Provenance: which CSV and which line. Makes any figure traceable back to
    -- a page of the paper book. Unique, so a row can never be imported twice.
    source_book INTEGER NOT NULL,
    source_row  INTEGER NOT NULL,
    UNIQUE (source_book, source_row)
);

-- Every change an app-entered flight has ever undergone, append-only.
--
-- The owner asked for a plain in-place edit and a real delete (2026-08-02), and
-- on a legal record that is only safe if the previous state survives somewhere:
-- a licence total that changes with no trace of what it was is exactly the
-- drift this project spent 106 KB of claude-docs/drift.md on. `before` is the
-- complete row as it stood, as JSON, so a delete is recoverable from this table
-- alone -- there is nowhere else left to read it from.
--
-- Nothing in the application reads this table and nothing anywhere updates or
-- deletes from it. It is written inside the same transaction as the change, so
-- a change without its audit entry cannot commit.
CREATE TABLE IF NOT EXISTS flight_audit (
    id      INTEGER PRIMARY KEY,
    at      TEXT    NOT NULL,
    user_id INTEGER NOT NULL,
    action  TEXT    NOT NULL CHECK (action IN ('update','delete')),
    -- Not a foreign key: the row it describes is gone in the delete case, and
    -- an audit trail that disappears with its subject is not an audit trail.
    seq     INTEGER NOT NULL,
    before  TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS flight_audit_by_seq ON flight_audit (seq);

CREATE INDEX IF NOT EXISTS flights_by_date  ON flights (flight_date);
CREATE INDEX IF NOT EXISTS flights_by_class ON flights (class);
CREATE INDEX IF NOT EXISTS flights_by_reg   ON flights (aircraft_reg);

-- What the source data says that the importer would not act on by itself.
-- Persisted so the app can show a review list, and so a later import that
-- resolves an item makes it disappear from the table rather than needing a
-- second place to be updated.
CREATE TABLE IF NOT EXISTS discrepancies (
    id          INTEGER PRIMARY KEY,
    kind        TEXT    NOT NULL,
    source_book INTEGER NOT NULL,
    source_row  INTEGER NOT NULL,
    flight_date TEXT    NOT NULL DEFAULT '',
    detail      TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS discrepancies_by_kind ON discrepancies (kind);

-- An append-only record of every import, so a suspect figure can be traced to
-- the run that produced it.
CREATE TABLE IF NOT EXISTS import_runs (
    id            INTEGER PRIMARY KEY,
    ran_at        TEXT    NOT NULL,
    flights       INTEGER NOT NULL,
    aircraft      INTEGER NOT NULL,
    discrepancies INTEGER NOT NULL,
    total_minutes INTEGER NOT NULL,
    landings      INTEGER NOT NULL,
    backup_path   TEXT    NOT NULL DEFAULT '',
    note          TEXT    NOT NULL DEFAULT ''
);

-- Auth. The logic lands with Task 4; the tables are declared here so the
-- schema is whole and the import's backup covers them from the start.
-- Design and reasoning: app/docs/security.md.
CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    -- Argon2id, with its parameters encoded in the hash so they can be raised
    -- later without invalidating existing hashes.
    password_hash TEXT NOT NULL,
    created_at    TEXT NOT NULL,
    disabled      INTEGER NOT NULL DEFAULT 0 CHECK (disabled IN (0,1))
);

CREATE TABLE IF NOT EXISTS sessions (
    id           INTEGER PRIMARY KEY,
    user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- The hash of the cookie's 256-bit identifier, never the identifier
    -- itself: reading this table must not yield usable session tokens.
    token_hash   TEXT NOT NULL UNIQUE,
    created_at   TEXT NOT NULL,
    last_used_at TEXT NOT NULL,
    -- Written on create and on each use, and NOT what decides anything: the
    -- window is store.SessionLifetime measured from last_used_at, so that
    -- changing it reaches the rows that already exist. This column is the
    -- value at rest, for anyone reading the file directly.
    expires_at   TEXT NOT NULL,
    user_agent   TEXT NOT NULL DEFAULT '',
    ip           TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS sessions_by_user ON sessions (user_id);
