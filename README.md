# Streaming Analytics With RisingWave and dbt
In today's data-driven world, the ability to harness real-time data for actionable insights is paramount. This repository serves as your gateway to mastering the art of streaming analytics using two powerful tools: RisingWave and dbt.

This repository is a comprehensive resource for hands-on demos, code examples, all aimed at guiding you through the exciting journey of building a robust streaming analytics framework.

![architecture.png](images%2Farchitecture.png)

# Clickstream analysis demo
Whether a website experiences daily traffic in the hundreds or millions, clickstream data offers valuable insights through various website metrics. This data not only provides insights into a page's popularity but also serves as an effective alternative to traditional market research methodologies. It allows us to assess the efficacy of digital marketing campaigns, UX design, and more by analyzing user behaviors, obviating the need for conventional surveys.

Clickstream data offers a detailed record of users' journeys through a website. It reveals the duration users spend on specific pages, their entry points, and subsequent interactions. If a significant number of users swiftly exit a page, it suggests the need for improvements to enhance user engagement.

In this tutorial, you will discover how to track the evolving click count of a webpage using RisingWave. We've established a demo cluster for your convenience, making it easy for you to experiment with these concepts.

![demo_architecture.png](images%2Fdemo_architecture.png)


## Connect RisingWave to data streams
After configuring the data stream in Redpanda with JSON format using the demonstration cluster, we can establish a connection to the streams using the following SQL statement. This data stream includes details about user interactions, including what each user is clicking on and the corresponding event timestamps.

```sql
CREATE SOURCE user_behaviors (
    user_id VARCHAR,
    target_id VARCHAR,
    target_type VARCHAR,
    event_timestamp TIMESTAMPTZ, --NOTE: Fix: change from TimeStamp
    behavior_type VARCHAR,
    parent_target_type VARCHAR,
    parent_target_id VARCHAR
) WITH (
    connector = 'kafka',
    topic = 'user_behaviors',
    properties.bootstrap.server = 'message_queue:29092',
    scan.startup.mode = 'earliest'
) FORMAT PLAIN ENCODE JSON;
```

## Create dbt project
### Install dbt-risingwave adapter
```shell
cd ./dbt
pip install -r requirements.txt
```

### Init dbt project
```shell
# create dbt project with risisingwave name
# input parameters to set up a profile to connect to the local RisingWave:
# - host: 127.0.0.1
# - port: 4566
# - user: root
dbt init risingwave_demo
# check
cd risingwave_demo
dbt debug --profiles-dir .
```
The content of the `profiles.yml` as below:
```yaml
risingwave_demo:
  outputs:
    dev:
      dbname: dev
      host: localhost
      password: ''
      port: 4566
      schema: public
      threads: 1
      type: risingwave
      user: root
  target: dev
```

### Creat a source
```yaml
sources:
  - name: dev_public_source
    database: dev
    schema: public
    tables:
      - name: user_behaviors
```

### Define materialized views and query the results
In this tutorial, we will create a materialized view that counts how many times a thread was clicked on over a day.

First, the tumble() function will map each event into a 10-minute window to create an intermediary table t, where the events will be aggregated on target_id and window_time to get the number of clicks for each thread. The events will also be filtered by target_type and behavior_type.

Next, the hop() function will create 24-hour time windows every 10 minutes. Each event will be mapped to corresponding windows. Finally, they will be grouped by target_id and window_time to calculate the total number of clicks of each thread within 24 hours.

The model should be like this:
```sql
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
            INTERVAL '1440 minutes' -- 24 hours
        )
GROUP BY
    target_id,
    window_start,
    window_end
```

### Query the results
You can use a SQL tool to query the most often viewed threads with the following statement.

```sql
SELECT * FROM thread_view_count
ORDER BY view_count DESC, window_start
LIMIT 10;
```
![results.png](images%2Fresults.png)

Or you can leverage the `dbt show` command:
```shell
dbt show --profiles-dir . --inline "select * from thread_view_count limit 10"

06:48:21  Running with dbt=1.5.6
06:48:22  Registered adapter: risingwave=1.5.1
06:48:26  [WARNING]: Configuration paths exist in your dbt_project.yml file which do not apply to any resources.
There are 1 unused configuration paths:
- models.risingwave_demo.example
06:48:26  Found 1 model, 0 tests, 0 snapshots, 0 analyses, 320 macros, 0 operations, 0 seed files, 1 source, 0 exposures, 0 metrics, 0 groups
06:48:26
06:48:26  Concurrency: 1 threads (target='dev')
06:48:26
06:48:27  Previewing inline node:
| target_id | view_count |         window_start |           window_end |
| --------- | ---------- | -------------------- | -------------------- |
| thread0   |        138 | 2023-09-19 21:20:... | 2023-09-20 21:20:... |
| thread0   |        185 | 2023-09-19 23:40:... | 2023-09-20 23:40:... |
| thread0   |        296 | 2023-09-20 04:40:... | 2023-09-21 04:40:... |
| thread0   |        308 | 2023-09-20 05:20:... | 2023-09-21 05:20:... |
| thread0   |        308 | 2023-09-20 05:30:... | 2023-09-21 05:30:... |
```

# Conclusion
In conclusion, the synergy between RisingWave and dbt presents an exciting opportunity to streamline the development of a robust and efficient streaming analytics platform.

RisingWave's simplicity and potential performance enhancements, when combined with dbt's powerful data transformation capabilities, create a compelling ecosystem. This partnership enables organizations to build a streaming analytics platform that not only processes data seamlessly but also allows for sophisticated data modeling, transformation, and analytics.

By leveraging RisingWave's strengths in stream processing and dbt's proficiency in data modeling, organizations can achieve a more agile and effective approach to real-time analytics. This combination empowers teams to rapidly develop, iterate, and deploy streaming analytics solutions while maintaining data quality and consistency.

In essence, the integration of RisingWave and dbt offers a holistic solution for organizations seeking to harness the power of real-time data analytics, making it easier and more effective to derive valuable insights from streaming data sources.

While RisingWave is touted for its potential performance improvements, it's important to note that we have not yet confirmed the extent of these enhancements. Initial indications suggest a promising outlook: stateless computing appears to offer a significant boost, ranging from 10% to 30%, while stateful computing shows the potential for an astonishing 10-fold or greater improvement.

To gain a more comprehensive understanding of RisingWave's performance capabilities, we eagerly anticipate the forthcoming performance report. Regardless, in the realm of stream processing, the combination of simplicity and performance remains a valuable asset, and RisingWave demonstrates its commitment to delivering on both fronts and more.

# Supporting Links
* <a href="https://www.risingwave.com/blog/rethinking-stream-processing-and-streaming-databases/" target="_blank">Rethinking Stream Processing and Streaming Databases</a>
* <a href="https://docs.risingwave.com/docs/current/use-dbt/" target="_blank">Use dbt for data transformations</a>
* <a href="https://www.risingwave.com/blog/rethinking-stream-processing-and-streaming-databases/" target="_blank"></a>
* <a href="https://docs.risingwave.com/docs/current/clickstream-analysis/" target="_blank">Clickstream analysis</a>
* <a href="https://www.clouddatainsights.com/real-time-olap-databases-and-streaming-databases-a-comparison/" target="_blank">Real-time OLAP Databases and Streaming Databases: A Comparison</a>
* <a href="https://www.risingwave.com/blog/start-your-stream-processing-journey-with-just-4-lines-of-code/" target="_blank">Start Your Stream Processing Journey With Just 4 Lines of Code</a>
* <a href="https://medium.com/@RisingWave_Engineering/top-8-streaming-databases-for-real-time-analytics-a-comprehensive-guide-f45d7b3b35c8" target="_blank">Top 8 Streaming Databases for Real-Time Analytics: A Comprehensive Guide</a>




