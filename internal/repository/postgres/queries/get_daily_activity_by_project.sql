WITH creations AS (
    SELECT
        DATE_TRUNC('day', created_time) AS day,
        COUNT(*) AS creation_count
    FROM raw.issue
    WHERE project_id = $1
    GROUP BY DATE_TRUNC('day', created_time)
),
completions AS (
    SELECT
        DATE_TRUNC('day', closed_time) AS day,
        COUNT(*) AS completion_count
    FROM raw.issue
    WHERE project_id = $1
        AND closed_time IS NOT NULL
    GROUP BY DATE_TRUNC('day', closed_time)
)
SELECT
    COALESCE(c.day, cm.day) AS date,
    COALESCE(c.creation_count, 0)::INT AS creation,
    COALESCE(cm.completion_count, 0)::INT AS completion
FROM creations c
FULL OUTER JOIN completions cm ON c.day = cm.day
ORDER BY COALESCE(c.day, cm.day)
