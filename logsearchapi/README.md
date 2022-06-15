# Log Search API Server for MinIO

## Development setup

1. Start Postgresql server in container:

```shell
docker run --rm -it -e "POSTGRES_PASSWORD=example" -p 5432:5432 postgres:13-alpine -c "log_statement=all"
```

2. Start logsearchapi server:

```shell
export LOGSEARCH_PG_CONN_STR="postgres://postgres:example@localhost/postgres"
export LOGSEARCH_AUDIT_AUTH_TOKEN=xxx
export MINIO_LOG_QUERY_AUTH_TOKEN=yyy
export LOGSEARCH_DISK_CAPACITY_GB=5
go build && ./logsearchapi
```

3. Minio setup:

```shell
mc admin config set myminio audit_webhook:1 'endpoint=http://localhost:8080/api/ingest?token=xxx'

mc admin service restart myminio
```

4. Sample search/list queries:

```shell
curl -v "http://localhost:8080/api/query?token=yyy&q=raw&pageNo=0&pageSize=10&timeStart=2020-11-04T22:26:12.732402319Z"
```

## Log Storage

Logs are stored in a PostgreSQL database, partitioned such that there are four tables for each month of data. When disk usage approaches the `LOGSEARCH_DISK_CAPACITY_GB` value, the oldest tables are automatically deleted so as to not run out of disk space.

Raw audit logs are stored as JSON columns. These tables can be queried by specifying the query parameter `q=raw`.

Additionally, a set of useful request parameters are extracted from the audit logs and stored in separate tables. These tables can be queried by specifying the query parameter `q=reqinfo`.

## API Documentation

### Ingest API

```
POST /api/ingest?token=xxx
```

This API is used to send MinIO audit logs for ingestion into the API service.

The `token` parameter is used to authenticate the request and should be equal to the `LOGSEARCH_AUDIT_AUTH_TOKEN` environment variable passed to the server.

The body must be a JSON object representing a single audit log object created by a MinIO server.

This endpoint must be configured as an audit log endpoint in the MinIO server.

#### Ingest API Filters

Specifying ingest API filters allows for including or excluding audit log storage based on pattern matching criteria.

Each API filter configuration is an include or exclude specification. If an incoming log matches any exclude configuration it is dropped without any further processing. If it does not match any exclude configuration, and if an include configuration is specified, the log is stored only if it matches an include configuration. If no include configuration is given, all logs not matching an exclude configuration are stored.

The following ingest filters are supported and are configured via the environment:

| Environment variable                      | Description                           |
|-------------------------------------------|---------------------------------------|
| `LOGSEARCH_INGEST_FILTER_APINAME_INCLUDE` | API Name matching criteria to include |
| `LOGSEARCH_INGEST_FILTER_APINAME_EXCLUDE` | API Name matching criteria to exclude |

The syntax for these environment variables is a list of [patterns](#pattern-matching) separated by semicolons (`;`). A log matches a filter if the concerned field matches any one of the patterns in the pattern list.

**LIMITATION**: A pattern MUST NOT contain `;`.

<details><summary>Example 1: Store only decommissioning logs and Delete operations</summary>

Set the following ingest filter environment variables:

```
LOGSEARCH_INGEST_FILTER_APINAME_INCLUDE='*Decom*;Delete*'
```
</details>

<details><summary>Example 2: Do not store Get operations</summary>

Set the following ingest filter environment variables:

```
LOGSEARCH_INGEST_FILTER_APINAME_EXCLUDE='Get*'
```
</details>

<details><summary>Example 3: Store only Put operations except PutBucket</summary>

Set the following ingest filter environment variables:

```
LOGSEARCH_INGEST_FILTER_APINAME_EXCLUDE='PutBucket'
LOGSEARCH_INGEST_FILTER_APINAME_INCLUDE='Put*'
```

</details>


### Query API

```
GET /api/query?token=xxx&...
```

This API is used to query MinIO audit logs stored by the API service.

The `token` parameter is used to authenticate the request and should be equal to the `MINIO_LOG_QUERY_AUTH_TOKEN` environment variable passed to the server.

Additional query parameters specify the logs to be retrieved and the format of their output.

| Query parameter      | Value Description                                                                                                                                                                        | Required | Default    |
|----------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|------------|
| `q`                  | `reqinfo` or `raw`.                                                                                                                                                                      | Yes      | -          |
| `timeStart`          | RFC3339 time or date. Examples: `2006-01-02T15:04:05.999999999Z07:00` or `2006-01-02`.                                                                                                   | No       | -          |
| `timeEnd`            | RFC3339 time or date. Examples: `2006-01-02T15:04:05.999999999Z07:00` or `2006-01-02`.                                                                                                   | No       | -          |
| `last`               | Represents a integer duration with unit (`24h` or `60m`). Use this to get logs for the most recent time window of the given length. Valid time units are "m" for minutes, "h" for hours. | No       | -          |
| `timeAsc`/`timeDesc` | Flag parameter (no value); either one may be specified. Specifies result ordering.                                                                                                       | No       | `timeDesc` |
| `fp`                 | Repeatable parameter specifying key-value match filters. See the [filter parameters](#filter-parameters) section.                                                                        | No       | -          |
| `pageSize`           | Number of results to return per API call. Allows values between 10 and 10000.                                                                                                            | No       | `10`       |
| `pageNo`             | 0-based page number of results.                                                                                                                                                          | No       | `0`        |
| `export`             | Specify an export format. This skips pagination. `csv` and `ndjson` are supported.                                                                                                       | No       | -          |

For example, to get the last 24 hours of request-info logs dumped in line-delimited JSON format:

```
# If your token contains URL-unsafe characters, it must be URL-encoded appropriately as shown with curl below:
curl -XGET -s \
   'http://logsearch:8080/api/query?q=reqinfo&timeAsc&export=ndjson&last=24h' \
   --data-urlencode 'token=xxx' > output.ndjson
```

When using an export format (csv/json), pagination parameters (`pageSize` and `pageNo`) are not used and all data matching data is returned.

#### Filter Parameters

Filter parameters allow filtering records based on pattern matching on the values of audit log fields. 

The format for each filter pattern is `key:value-pattern`.

Allowed values for the `key` are:

| Valid Keys        |
|-------------------|
| `bucket`          |
| `object`          |
| `api_name`        |
| `request_id`      |
| `user_agent`      |
| `response_status` |

Pattern matching is defined [here](#pattern-matching). As an example `bucket:photos-*` matches any bucket with a `photos-` prefix. 

When multiple filter parameters are given, only records matching all filter parameters are returned.

<details><summary>Example 1: Filter and export request info logs of Put operations on the bucket `photos` in last 24 hours</summary>

```
curl -XGET -s \
   'http://logsearch:8080/api/query?q=reqinfo&timeAsc&export=ndjson&last=24h&fp=bucket:photos&fp=api_name:Put*' \
   --data-urlencode 'token=xxx' > output.ndjson
```

</details>

<details><summary>Example 2: Filter for the first 1000 raw audit logs of decommissioning operations in the last 1000 hours</summary>

```
curl -XGET -s \
   'http://logsearch:8080/api/query?q=raw&timeAsc&pageSize=1000&last=1000h&fp=api_name:*Decom*' \
   --data-urlencode 'token=xxx'
```

</details>

### Pattern matching

LogsearchAPI's support for pattern matching is specified here. The pattern matching supported is the same as database style wilcard (or glob) pattern matching. A pattern string without the `.` or `*` characters matches only itself. A `.` matches any single character and a `*` matches 0 or more characters. To match a literal `.` or `*` prefix it with a `\`. To match a literal `\`, just double it: `\\`. The value pattern is case-sensitive.

For example `photos-*` matches any string starting with `photos-`.
