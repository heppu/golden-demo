CREATE TYPE task_status AS ENUM ('waiting', 'working', 'done');

CREATE TABLE
    tasks (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        description TEXT,
        status task_status NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
