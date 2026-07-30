-- +goose Up
-- +goose StatementBegin
CREATE TABLE users(
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    hash_password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()

);

CREATE TABLE languages(
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    slug VARCHAR(30) NOT NULL UNIQUE,
    build_command TEXT,
    execute_command TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    language_id INT NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    time_limit_ms INT NOT NULL default 1000,
    memory_limit_mb INT NOT NULL default 256,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE test_cases(
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    input_data TEXT NOT NULL,
    output_data TEXT NOT NULL,
    is_hidden BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE submissions(
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    language_id INT NOT NULL REFERENCES languages(id) ON DELETE RESTRICT,
    code TEXT NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'In Queue',
    execution_time_ms INT,
    memory_used_mb INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS test_cases;
DROP TABLE IF EXISTS languages;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
