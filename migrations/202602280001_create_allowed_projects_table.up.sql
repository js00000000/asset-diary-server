CREATE TABLE IF NOT EXISTS allowed_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert 'read-together' as the initial allowed project
INSERT INTO allowed_projects (name) VALUES ('read-together');

-- Update waiting_lists to use foreign key (optional but good for data integrity)
-- Since waiting_lists already has 'project' column as VARCHAR, let's make it reference allowed_projects(name)
ALTER TABLE waiting_lists 
ADD CONSTRAINT fk_waiting_lists_project 
FOREIGN KEY (project) REFERENCES allowed_projects(name) ON UPDATE CASCADE;
