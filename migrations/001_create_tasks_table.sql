-- migrations/001_create_tasks_table.sql
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    done BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insertar datos de ejemplo
INSERT INTO tasks (title, description) VALUES
    ('Tarea 1', 'Descripción de tarea 1'),
    ('Tarea 2', 'Descripción de tarea 2');
