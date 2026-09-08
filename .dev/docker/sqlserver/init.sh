#!/usr/bin/env bash
set -euo pipefail

if [[ ! "${SQLSERVER_DATABASE}" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]]; then
    echo "Invalid SQLSERVER_DATABASE: ${SQLSERVER_DATABASE}" >&2
    exit 1
fi

/opt/mssql-tools18/bin/sqlcmd \
    -C \
    -S sqlserver \
    -U sa \
    -P "${SQLSERVER_SA_PASSWORD}" \
    -v "DatabaseName=${SQLSERVER_DATABASE}" \
    -i /usr/local/share/prismgo/init.sql \
    -b
