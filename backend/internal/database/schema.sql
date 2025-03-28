-- Create nutritional_info table
CREATE TABLE IF NOT EXISTS nutritional_info (
    id TEXT PRIMARY KEY,
    total_weight REAL NOT NULL,
    calories REAL NOT NULL,
    protein REAL NOT NULL,
    carbs REAL NOT NULL,
    fat REAL NOT NULL,
    fiber REAL NOT NULL,
    sugar REAL NOT NULL,
    image_path TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Create nutrition_scans table
CREATE TABLE IF NOT EXISTS nutrition_scans (
    id TEXT PRIMARY KEY,
    image_data BLOB NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('pending', 'processing', 'completed', 'failed')),
    error TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_nutrition_scans_status ON nutrition_scans(status); 