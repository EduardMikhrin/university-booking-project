-- Add photo column to users table with default value
ALTER TABLE users
ADD COLUMN IF NOT EXISTS photo VARCHAR(500) DEFAULT 'https://cdn-icons-png.flaticon.com/512/709/709699.png';

-- Add comment to photo column
COMMENT ON COLUMN users.photo IS 'URL or path to user profile photo';

-- Update existing users to have the default photo if they don't have one
UPDATE users
SET photo = 'https://cdn-icons-png.flaticon.com/512/709/709699.png'
WHERE photo IS NULL;

