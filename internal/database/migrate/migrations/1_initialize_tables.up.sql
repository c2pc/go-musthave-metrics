CREATE TABLE IF NOT EXISTS gauges
(
    key   VARCHAR(64) UNIQUE NOT NULL,
    value DOUBLE PRECISION   NOT NULL
);

CREATE TABLE IF NOT EXISTS counters
(
    key   VARCHAR(64) UNIQUE NOT NULL,
    value BIGINT   NOT NULL
);