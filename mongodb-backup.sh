DB_CONTAINER="your_container_name"; \
DB_NAME="your_database_name"; \
MONGO_USER="admin"; \
MONGO_PASS="pass"; \
MONGO_AUTH_DB="admin"; \
HOST_BACKUP_DIR="mongo_dump"; \
CONTAINER_DUMP_DIR="/tmp/db_dump"; \
\
mkdir -p "${HOST_BACKUP_DIR}" && \
\
# 1. 在容器内执行 mongodump，生成到 /tmp/db_dump
docker exec "${DB_CONTAINER}" mongodump \
    --authenticationDatabase="${MONGO_AUTH_DB}" \
    --db="${DB_NAME}" \
    --username="${MONGO_USER}" \
    --password="${MONGO_PASS}" \
    --out="${CONTAINER_DUMP_DIR}" && \
\
# 2. 将容器内的备份文件拷贝到宿主机
docker cp "${DB_CONTAINER}:${CONTAINER_DUMP_DIR}" "${HOST_BACKUP_DIR}" && \
\
# 3. 清理容器内生成的临时文件
docker exec "${DB_CONTAINER}" rm -rf "${CONTAINER_DUMP_DIR}" && \
\
echo "数据库 ${DB_NAME} 已成功备份到宿主机: ${HOST_BACKUP_DIR}"