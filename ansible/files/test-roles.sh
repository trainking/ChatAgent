#!/bin/bash
ROOT=$(curl -s -X POST http://localhost/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"root@chatme365.online","password":"Cha@365mexXy"}')
TOKEN=$(echo "$ROOT" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

echo '=== All Permissions ==='
curl -s http://localhost/api/v1/permissions -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
for p in json.load(sys.stdin)['data']:
    print(f'  {p[\"code\"]:20s} {p[\"name\"]}')
"

echo '=== Admin Permissions ==='
curl -s http://localhost/api/v1/roles/admin/permissions -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)['data']
print(' ', d['permissions'])
"

echo '=== Agent Permissions ==='
curl -s http://localhost/api/v1/roles/agent/permissions -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)['data']
print(' ', d['permissions'])
"

echo '=== Update Admin Permissions (add roles.manage) ==='
curl -s -X PUT http://localhost/api/v1/roles/admin/permissions \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"permissions":["inbox.view","inbox.reply","inbox.assign","users.manage","reports.view","roles.manage"]}'
echo

echo '=== Admin Not SuperAdmin (forbidden on roles) ==='
ADMIN=$(curl -s -X POST http://localhost/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin1@chatme365.online","password":"Ad@123Mxm"}')
ATOKEN=$(echo "$ADMIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
curl -s http://localhost/api/v1/roles/admin/permissions -H "Authorization: Bearer $ATOKEN"
echo
