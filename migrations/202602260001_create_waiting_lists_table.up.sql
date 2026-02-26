CREATE TABLE IF NOT EXISTS waiting_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    project VARCHAR(100) NOT NULL,
    information JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(email, project)
);

CREATE INDEX IF NOT EXISTS idx_waiting_lists_project ON waiting_lists(project);
CREATE INDEX IF NOT EXISTS idx_waiting_lists_email ON waiting_lists(email);
