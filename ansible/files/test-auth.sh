#!/bin/bash
echo '=== Status (should be uninitialized) ==='
curl -s http://localhost/api/v1/auth/status
echo

echo '=== Init Root ==='
INIT=$(curl -s -X POST http://localhost/api/v1/auth/init -H 'Content-Type: application/json' -d '{"email":"root@chatagent.com","password":"root123","name":"RootAdmin"}')
echo "$INIT"
echo

echo '=== Init Again (should fail) ==='
curl -s -X POST http://localhost/api/v1/auth/init -H 'Content-Type: application/json' -d '{"email":"xxx@test.com","password":"test123","name":"Test"}'
echo

echo '=== Status (should be initialized) ==='
curl -s http://localhost/api/v1/auth/status
echo

echo '=== Login Root ==='
LOGIN=$(curl -s -X POST http://localhost/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"root@chatagent.com","password":"root123"}')
echo "$LOGIN" | python3 -c "import sys,json; d=json.load(sys.stdin); print('must_change_password:', d['data']['user']['must_change_password'])"

echo '=== Admin Creates Agent ==='
TOKEN=$(echo "$LOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
curl -s -X POST http://localhost/api/v1/users -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"email":"agent1@chatagent.com","name":"AgentOne","password":"agent123","role":"agent"}'
echo

echo '=== Login Agent (must_change_password should be true) ==='
ALOGIN=$(curl -s -X POST http://localhost/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"agent1@chatagent.com","password":"agent123"}')
echo "$ALOGIN" | python3 -c "import sys,json; d=json.load(sys.stdin); print('must_change_password:', d['data']['user']['must_change_password'])"

echo '=== Agent Changes Password ==='
ATOKEN=$(echo "$ALOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
curl -s -X POST http://localhost/api/v1/auth/change-password -H "Authorization: Bearer $ATOKEN" -H 'Content-Type: application/json' -d '{"old_password":"agent123","new_password":"newpass123"}'
echo

echo '=== Agent Login After Change ==='
ALOGIN2=$(curl -s -X POST http://localhost/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"agent1@chatagent.com","password":"newpass123"}')
echo "$ALOGIN2" | python3 -c "import sys,json; d=json.load(sys.stdin); print('must_change_password:', d['data']['user']['must_change_password'])"

echo '=== Agent Access Users (should fail, not admin) ==='
ATOKEN2=$(echo "$ALOGIN2" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
curl -s http://localhost/api/v1/users -H "Authorization: Bearer $ATOKEN2"
echo
