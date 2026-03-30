SELECT COALESCE(MAX(updated_time), '1970-01-01'::TIMESTAMP)
FROM raw.issue
WHERE project_id = $1
