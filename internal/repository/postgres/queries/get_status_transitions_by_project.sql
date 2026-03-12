SELECT sc.change_time, sc.from_status, sc.to_status
FROM raw.status_changes sc
JOIN raw.issue i ON i.id = sc.issue_id
WHERE i.project_id = $1
    AND i.closed_time IS NOT NULL
ORDER BY sc.change_time
