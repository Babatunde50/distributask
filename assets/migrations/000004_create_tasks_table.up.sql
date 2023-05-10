CREATE TYPE task_status AS ENUM (
    'queued',
    'in_progress',
    'completed',
    'failed'
);
CREATE TYPE task_type AS ENUM ('image_processing');


CREATE TABLE tasks (
    id SERIAL NOT NULL PRIMARY KEY,
    type task_type NOT NULL,
    payload JSONB NOT NULL,
    priority INTEGER NOT NULL DEFAULT 1,
    status task_status NOT NULL DEFAULT 'queued',
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    timeout INTEGER NOT NULL DEFAULT 30,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 5,
    user_id INTEGER NOT NULL REFERENCES users(id),
    result TEXT NOT NULL DEFAULT '-'
);