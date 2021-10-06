BEGIN;

ALTER TABLE employees
    DROP COLUMN phone,
    DROP COLUMN email;

COMMIT;