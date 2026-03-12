WITH issue_stats AS (
    SELECT
        COUNT(*) AS total,
        COUNT(*) FILTER (WHERE status = 'Open') AS count_open,
        COUNT(*) FILTER (WHERE status = 'Closed') AS count_closed,
        COUNT(*) FILTER (WHERE status = 'Resolved') AS count_resolved,
        COUNT(*) FILTER (WHERE status = 'In Progress') AS count_in_progress,
        COALESCE(SUM(EXTRACT(EPOCH FROM (closed_time - created_time))), 0)::INT AS total_duration_closed,
        COUNT(*) FILTER (WHERE created_time >= NOW() - INTERVAL '7 days') AS count_created_last_week
    FROM raw.issue
    WHERE project_id = $1
),
reopened_count AS (
    SELECT COUNT(DISTINCT sc.issue_id) AS count_reopened
    FROM raw.status_changes sc
    JOIN raw.issue i ON sc.issue_id = i.id AND i.project_id = $1
    WHERE sc.from_status IN ('Closed', 'Resolved', 'Done')
        AND sc.to_status IN ('Open', 'In Progress')
)
SELECT
    COALESCE(i.total, 0),
    COALESCE(i.count_open, 0),
    COALESCE(i.count_closed, 0),
    COALESCE(r.count_reopened, 0),
    COALESCE(i.count_resolved, 0),
    COALESCE(i.count_in_progress, 0),
    COALESCE(i.total_duration_closed, 0),
    COALESCE(i.count_created_last_week, 0)
FROM issue_stats i
CROSS JOIN reopened_count r
