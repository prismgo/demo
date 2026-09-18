#!/bin/sh
# The official MySQL image grants MYSQL_USER access only to MYSQL_DATABASE.
# Grant the schema demo a disposable schema so it can exercise CREATE/DROP
# DATABASE locally. Fixed database name, configurable user.
set -e

if [ -z "${MYSQL_USER}" ] || [ -z "${MYSQL_ROOT_PASSWORD}" ]; then
  echo "mysql init: MYSQL_USER or MYSQL_ROOT_PASSWORD is empty; skipping schema demo grant" >&2
  exit 0
fi

mysql --protocol=socket -uroot -p"${MYSQL_ROOT_PASSWORD}" <<SQL
GRANT ALL PRIVILEGES ON \`prismgo_schema_demo_test\`.* TO '${MYSQL_USER}'@'%';
FLUSH PRIVILEGES;
SQL
