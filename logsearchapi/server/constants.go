// Copyright (C) 2020, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package server

// Environment variable name constants
const (
	QueryAuthTokenEnv    = "MINIO_LOG_QUERY_AUTH_TOKEN"
	PgConnStrEnv         = "LOGSEARCH_PG_CONN_STR"
	AuditAuthTokenEnv    = "LOGSEARCH_AUDIT_AUTH_TOKEN"
	DiskCapacityEnv      = "LOGSEARCH_DISK_CAPACITY_GB"
	APINameIncludeFilter = "LOGSEARCH_INGEST_FILTER_APINAME_INCLUDE"
	APINameExcludeFilter = "LOGSEARCH_INGEST_FILTER_APINAME_EXCLUDE"
)
