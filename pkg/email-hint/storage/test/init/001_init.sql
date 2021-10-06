CREATE USER gopher
WITH PASSWORD 'P@ssw0rd';

CREATE DATABASE gopher_corp
    WITH OWNER gopher
    TEMPLATE = 'template0'
    ENCODING = 'utf-8'
    LC_COLLATE = 'C.UTF-8'
    LC_CTYPE = 'C.UTF-8';