BEGIN;

ALTER TABLE employees
    ADD COLUMN phone text,
    ADD COLUMN email text;

COMMIT;