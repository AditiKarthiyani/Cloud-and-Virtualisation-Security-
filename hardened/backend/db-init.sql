CREATE TABLE IF NOT EXISTS todos (
  id serial PRIMARY KEY,
  title text NOT NULL,
  done boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO todos (title, done)
SELECT 'Welcome to your todo list', false
WHERE NOT EXISTS (SELECT 1 FROM todos);
