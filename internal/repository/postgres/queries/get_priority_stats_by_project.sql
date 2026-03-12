SELECT
    priority,
    COUNT(*)::INT AS count
FROM raw.issue
WHERE project_id = $1
GROUP BY priority
ORDER BY priority
