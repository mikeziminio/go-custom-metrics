CREATE TYPE metric_type AS ENUM ('counter', 'gauge');

CREATE TABLE metric (
    id VARCHAR PRIMARY KEY,
    m_type metric_type NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION
);
