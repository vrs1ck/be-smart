CREATE TABLE IF NOT EXISTS gocourse.expenses (
    id          SERIAL PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    amount      NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    category    VARCHAR(50) NOT NULL DEFAULT '',
    covered_by  VARCHAR(50) NOT NULL DEFAULT '',
    month       INT NOT NULL CHECK (month BETWEEN 1 AND 12),
    year        INT NOT NULL CHECK (year >= 2020),
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

CREATE INDEX ON gocourse.expenses(year, month);
