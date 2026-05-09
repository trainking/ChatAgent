#!/bin/bash
echo '=== Login as root ==='
LOGIN=$(curl -s -X POST http://localhost/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"root@chatme365.online","password":"Cha@365mexXy"}')
TOKEN=$(echo "$LOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

echo '=== SuperAdmin resets agent password ==='
curl -s -X PUT http://localhost/api/v1/users/`docker exec deploy-postgres-1 psql -U postgres -d chatagent -tAc "SELECT id FROM users WHERE email='agent1@chatme365.online'"`/reset-password \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json'
echo

echo '=== Admin login ==='
ALOGIN=$(curl -s -X POST http://localhost/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin1@chatme365.online","password":"Ad@123Mxm"}')
ATOKEN=$(echo "$ALOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

echo '=== Admin tries reset (should fail: forbidden) ==='
AGENT_ID=$(docker exec deploy-postgres-1 psql -U postgres -d chatagent -tAc "SELECT id FROM users WHERE email='agent1@chatme365.online'")
curl -s -X PUT http://localhost/api/v1/users/$AGENT_ID/reset-password \
  -H "Authorization: Bearer $ATOKEN" \
  -H 'Content-Type: application/json'
echo
