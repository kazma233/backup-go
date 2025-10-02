DB_CONTAINER="your_container_name"; \
DB_NAME="your_database_name"; \
MONGO_USER="admin"; \
MONGO_PASS="pass"; \
MONGO_AUTH_DB="admin"; \
HOST_BACKUP_DIR="mongo_dump"; \
HOST_BACKUP_FILE="${HOST_BACKUP_DIR}/db_$(date +%Y%m%d_%H%M%S).tgz"; \
CONTAINER_DUMP_FILE="/tmp/db_dump.tgz"; \
\
mkdir -p "${HOST_BACKUP_DIR}" && \
\
# 1. 在容器内执行 mongodump，生成压缩文件到 /tmp/db_dump.tgz
docker exec "${DB_CONTAINER}" mongodump \
    --authenticationDatabase="${MONGO_AUTH_DB}" \
    --db="${DB_NAME}" \
    --username="${MONGO_USER}" \
    --password="${MONGO_PASS}" \
    --gzip \
    --archive="${CONTAINER_DUMP_FILE}" && \
\
# 2. 将容器内的备份文件拷贝到宿主机
docker cp "${DB_CONTAINER}:${CONTAINER_DUMP_FILE}" "${HOST_BACKUP_FILE}" && \
\
# 3. 清理容器内生成的临时文件
docker exec "${DB_CONTAINER}" rm "${CONTAINER_DUMP_FILE}" && \
\
echo "数据库 ${DB_NAME} 已成功备份到宿主机文件: ${HOST_BACKUP_FILE}"