DB_CONTAINER="your_container_name"; \
DB_NAME="your_database_name"; \
DB_USER="user"; \
HOST_BACKUP_DIR="postgres_dump"; \
HOST_BACKUP_FILE="${HOST_BACKUP_DIR}/db_$(date +%Y%m%d_%H%M%S).dump"; \
mkdir -p ${HOST_BACKUP_DIR} && \
\
# 在 Linux 环境下，当客户端（pg_dump）和服务器（postgres 进程）位于同一台机器或同一容器内时，它们会优先通过 Unix Domain Socket 进行通信，而不是走标准的 TCP/IP 端口
docker exec "${DB_CONTAINER}" pg_dump -U "${DB_USER}" -d "${DB_NAME}" -Fc > "${HOST_BACKUP_FILE}"