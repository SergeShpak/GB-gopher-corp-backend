CREATE TABLE departments_budget (
    id INT PRIMARY KEY REFERENCES departments(id),
    budget MONEY
);