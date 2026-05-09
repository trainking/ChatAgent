#!/bin/bash
echo '=== Init Root ==='
INIT=$(curl -s -X POST http://localhost/api/v1/auth/init -H 'Content-Type: application/json' -d '{"email":"root@chatagent.com","password":"root123","name":"SuperAdmin"}')
echo "$INIT" | python3 -c "import sys,json; d=json.load(sys.stdin); print('role:', d['data']['user']['role'])"
ROOT_TOKEN=$(echo "$INIT" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

echo '=== Root Creates Admin ==='
curl -s -X POST http://localhost/api/v1/users -H "Authorization: Bearer $ROOT_TOKEN" -H 'Content-Type: application/json' -d '{"email":"admin@chatagent.com","name":"Admin","password":"admin123","role":"admin"}'
echo

echo '=== Root Creates Agent ==='
curl -s -X POST http://localhost/api/v1/users -H "Authorization: Bearer $ROOT_TOKEN" -H 'Content-Type: application/json' -d '{"email":"agent@chatagent.com","name":"Agent","password":"agent123","role":"agent"}'
echo

echo '=== Root sees all users (including super_admin) ==='
ROOT_USERS=$(curl -s http://localhost/api/v1/users -H "Authorization: Bearer $ROOT_TOKEN")
echo "$ROOT_USERS" | python3 -c "import sys,json; [print(u['email'], u['role']) for u in json.load(sys.stdin)['data']]"

echo '=== Admin Login ==='
ALOGIN=$(curl -s -X POST http://localhost/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"admin@chatagent.com","password":"admin123"}')
ADMIN_TOKEN=$(echo "$ALOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

echo '=== Admin sees users (NO super_admin) ==='
ADMIN_USERS=$(curl -s http://localhost/api/v1/users -H "Authorization: Bearer $ADMIN_TOKEN")
echo "$ADMIN_USERS" | python3 -c "import sys,json; [print(u['email'], u['role']) for u in json.load(sys.stdin)['data']]"
