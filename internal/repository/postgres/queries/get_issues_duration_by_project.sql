SELECT
    id,
    EXTRACT(EPOCH FROM (closed_time - created_time))::INT AS duration_seconds
FROM raw.issue
WHERE project_id = $1
    AND closed_time IS NOT NULL
ORDER BY id
