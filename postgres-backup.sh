DB_CONTAINER="your_container_name"; \
 DB_NAME="your_database_name"; \
 DB_USER="user"; \
 HOST_BACKUP_DIR="postgres_dump"; \
 HOST_BACKUP_FILE="db_$(date +%Y%m%d_%H%M%S).dump"; \
 mkdir -p ${HOST_BACKUP_DIR} && \
 docker exec "${DB_CONTAINER}" pg_dump -U "${DB_USER}" -d "${DB_NAME}" -Fc > "${HOST_BACKUP_DIR}/${HOST_BACKUP_FILE}"