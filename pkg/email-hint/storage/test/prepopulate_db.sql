BEGIN;

INSERT INTO positions (title)
VALUES
    ('CTO'),
    ('CEO'),
    ('CSO'),
    ('Backend Dev'),
    ('Frontend Dev'),
    ('Fullstack Dev'),
    ('QA'),
    ('Technical writer')
ON CONFLICT(title) DO NOTHING;

INSERT INTO departments (id, parent_id, name)
OVERRIDING SYSTEM VALUE
VALUES
    (0, 0, 'root') ON CONFLICT(id) DO NOTHING;

INSERT INTO departments (parent_id, name)
VALUES
    (
        (
            SELECT id
            FROM departments
            WHERE name = 'root'
        ),
        'executives'
    ),
    (
        (
            SELECT id
            FROM departments
            WHERE name = 'root'
        ),
        'R&D'
    ),
    (
        (
            SELECT id
            FROM departments
            WHERE name = 'root'
        ),
        'Accounting'
    ),
    (
        (
            SELECT id
            FROM departments
            WHERE name = 'root'
        ),
        'Sales'
    ) ON CONFLICT(name) DO NOTHING;

COMMIT;

BEGIN DEFERRABLE;
    INSERT INTO employees (first_name, last_name, phone, email, salary, manager_id, department, position)
    VALUES
        (
            'Bob',
            'Morane',
            '+79231234567',
            'bmorane@gopher_corp.com',
            500000,
            42,
            (SELECT id FROM departments WHERE name = 'executives'),
            (SELECT id FROM positions WHERE title = 'CSO')
        );
    UPDATE employees
    SET manager_id = (
        SELECT id
        FROM employees
        WHERE
            first_name = 'Bob'
            AND last_name = 'Morane'
        LIMIT 1
    )
    WHERE
        first_name = 'Bob'
        AND last_name = 'Morane';

    INSERT INTO employees (first_name, last_name, phone, email, salary, manager_id, department, position)
    VALUES
        (
            'Charley',
            'Bucket',
            '+79159876543',
            'cbucket@gopher_corp.com',
            1000000,
            42,
            (SELECT id FROM departments WHERE name = 'executives'),
            (SELECT id FROM positions WHERE title = 'CEO')
        );
    UPDATE employees
    SET manager_id = (
        SELECT id
        FROM employees
        WHERE
            first_name = 'Charley'
            AND last_name = 'Bucket'
        LIMIT 1
    )
    WHERE
        first_name = 'Charley'
        AND last_name = 'Bucket';

    INSERT INTO employees (first_name, last_name, phone, email, salary, manager_id, department, position)
    VALUES
        (
            'Alice',
            'Liddell',
            '+79169008070',
            'aliddell@gopher_corp.com',
            500000,
            42,
            (SELECT id FROM departments WHERE name = 'executives'),
            (SELECT id FROM positions WHERE title = 'CTO')
        );
    UPDATE employees
    SET manager_id = (
        SELECT id
        FROM employees
        WHERE
            first_name = 'Alice'
            AND last_name = 'Liddell'
        LIMIT 1
    )
    WHERE
        first_name = 'Alice'
        AND last_name = 'Liddell';
COMMIT;