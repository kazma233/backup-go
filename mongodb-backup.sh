DB_CONTAINER="your_container_name"; \
 DB_NAME="your_database_name"; \
 MONGO_USER="admin"; \
 MONGO_PASS="pass"; \
 MONGO_AUTH_DB="admin";
 HOST_BACKUP_DIR="mongo_dump"; \
 HOST_BACKUP_FILE="db_$(date +%Y%m%d_%H%M%S).tgz"; \
 mkdir -p ${HOST_BACKUP_DIR} && \
 docker exec "${MONGO_CONTAINER}" mongodump --authenticationDatabase="${MONGO_AUTH_DB}" --db="${DB_NAME}" --username="${MONGO_USER}" --password="${MONGO_PASS}" --gzip > "${HOST_BACKUP_DIR}/${HOST_BACKUP_FILE}"