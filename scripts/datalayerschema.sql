DO $$
BEGIN
    IF NOT EXISTS (
        SELECT FROM pg_database WHERE datname = 'golangtodo'
    ) THEN
        CREATE DATABASE golangtodo;
    END IF;

    IF NOT EXISTS (
        SELECT FROM pg_roles WHERE rolname = 'goadmin'
    ) THEN
        CREATE USER goadmin WITH PASSWORD 'your_password';
    END IF;

    GRANT ALL PRIVILEGES ON DATABASE golangtodo TO goadmin;
END
$$;