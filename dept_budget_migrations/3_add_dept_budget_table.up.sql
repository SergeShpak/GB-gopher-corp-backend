CREATE TABLE departments_budget (
    department INT UNIQUE REFERENCES departments(id),
    budget MONEY
);