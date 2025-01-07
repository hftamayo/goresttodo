DO $$
BEGIN
    IF NOT EXISTS (
        SELECT FROM pg_database WHERE datname = 'golangtodo'
    ) THEN
        CREATE DATABASE golangtodo;
    END IF;

    IF NOT EXISTS (
        SELECT FROM pg_roles WHERE rolname = 'sebastic'
    ) THEN
        CREATE USER sebastic WITH PASSWORD 'milucito';
    END IF;

    GRANT ALL PRIVILEGES ON DATABASE golangtodo TO sebastic;
END
$$;
