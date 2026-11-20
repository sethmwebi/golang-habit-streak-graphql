CREATE TABLE IF NOT EXISTS habit_logs(
  id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  habit_id TEXT NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
  completed_date DATE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(habit_id, completed_date)
);

CREATE INDEX IF NOT EXISTS idx_habit_logs_completed_date ON habit_logs(completed_date);
