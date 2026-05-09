#!/bin/bash
# 恢复系统到初始状态并写入三个测试账号

echo '=== 清空数据库 ==='
docker exec deploy-postgres-1 psql -U postgres -d chatagent -c 'DELETE FROM users;'

echo '=== 创建根账号 (超级管理员) ==='
INIT=$(curl -s -X POST http://localhost/api/v1/auth/init \
  -H 'Content-Type: application/json' \
  -d '{"email":"root@chatme365.online","password":"Cha@365mexXy","name":"RootAdmin"}')
echo "$INIT" | python3 -c "import sys,json; d=json.load(sys.stdin); print('根账号创建:', d['code'] == 0 and 'OK' or 'FAIL')"
ROOT_TOKEN=$(echo "$INIT" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

echo '=== 创建管理员 ==='
curl -s -X POST http://localhost/api/v1/users \
  -H "Authorization: Bearer $ROOT_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin1@chatme365.online","name":"Admin","password":"Ad@123Mxm","role":"admin"}'
echo

echo '=== 创建客服 ==='
curl -s -X POST http://localhost/api/v1/users \
  -H "Authorization: Bearer $ROOT_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"email":"agent1@chatme365.online","name":"Agent","password":"Ag&1234Bbo","role":"agent"}'
echo

echo '=== 验证三个账号 ==='
curl -s http://localhost/api/v1/users -H "Authorization: Bearer $ROOT_TOKEN" | python3 -c "
import sys,json
for u in json.load(sys.stdin)['data']:
    print(f'{u[\"email\"]:30s} {u[\"role\"]}')
"

echo '=== 完成 ==='
