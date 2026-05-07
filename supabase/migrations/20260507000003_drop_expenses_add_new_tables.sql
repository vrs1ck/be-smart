DROP TABLE IF EXISTS gocourse.expenses CASCADE;

CREATE TABLE IF NOT EXISTS gocourse.expenses (
    id         SERIAL PRIMARY KEY,
    title      VARCHAR(255) NOT NULL,
    budget     NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (budget >= 0),
    recurring  BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- transactions records each actual payment linked to an expense.
-- ON DELETE CASCADE: deleting an expense automatically deletes all its transactions.
CREATE TABLE IF NOT EXISTS gocourse.transactions (
    id         SERIAL PRIMARY KEY,
    expense_id INT NOT NULL REFERENCES gocourse.expenses(id) ON DELETE CASCADE,
    paid_by    VARCHAR(100) NOT NULL,
    amount     NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    comment    TEXT NOT NULL DEFAULT '',
    month      INT NOT NULL CHECK (month BETWEEN 1 AND 12),
    year       INT NOT NULL CHECK (year >= 2020),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX ON gocourse.transactions(expense_id);
CREATE INDEX ON gocourse.transactions(month, year);
