{{ config(materialized='materializedview') }}

WITH t AS (
    SELECT
        target_id,
        COUNT() AS view_count,
        window_start AS window_time
    FROM
        TUMBLE(
            {{ source('dev_public_source', 'user_behaviors') }},
            event_timestamp,
            INTERVAL '10 minutes'
        )
    WHERE
        target_type = 'thread'
        AND behavior_type = 'show'
    GROUP BY
        target_id,
        window_start
)
SELECT
    target_id,
    SUM(t.view_count) AS view_count,
    window_start,
    window_end
FROM
    HOP(
            t,
            t.window_time,
            INTERVAL '10 minutes',
            INTERVAL '1440 minutes'
        )
GROUP BY
    target_id,
    window_start,
    window_end