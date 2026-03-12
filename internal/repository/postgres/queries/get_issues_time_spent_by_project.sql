SELECT
    id,
    time_spent
FROM raw.issue
WHERE project_id = $1
    AND closed_time IS NOT NULL
    AND time_spent IS NOT NULL
ORDER BY id
