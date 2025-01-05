# backup-go

backup your dir 2 oss, support tg/mail notice.

config example (at: config/config.yml):
``` yml
mail:
  smtp: 'smtp'
  port: 'port'
  user: 'user'
  password: 'password'
notice_mail: 
  - 'notice_mail'
tg:
  key: 'key'
tg_chat_id: '@tg_chat_id'

# must
oss:
  bucket_name: 'bucket_name'
  endpoint: 'endpoint'
  fast_endpoint: 'fast_endpoint'
  access_key: 'access_key'
  access_key_secret: 'access_key_secret'
backup:
  # support multiple
  app1:
    # zip before command
    before_command: 'docker cp xxx:/app/data/ ./export'
    # zip dir, must setting
    back_path: './export'
    # zip after command
    after_command: 'rm -rf ./export'
    # backup cron
    backup_task: '0 25 0 * * ?'
    # liveness cron check task availability
    liveness: '0 0 0 * * ?'
  app2:
    back_path: './export'
    backup_task: '0 25 0 * * ?'
    liveness: '0 0 0 * * ?'
```

run:
``` shell
chmod +x rebuild.sh
./rebuild.sh
```