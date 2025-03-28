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

    IF NOT EXISTS (
        SELECT FROM pg_database WHERE datname = 'nodetodo'
    ) THEN
        CREATE DATABASE nodetodo;
    END IF;

    GRANT ALL PRIVILEGES ON DATABASE golangtodo TO sebastic;
    GRANT ALL PRIVILEGES ON DATABASE nodetodo TO sebastic;
END
$$;
