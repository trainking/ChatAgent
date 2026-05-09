#!/bin/bash
ROOT=$(curl -s -X POST http://localhost/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"root@chatme365.online","password":"Cha@365mexXy"}')
TOKEN=$(echo "$ROOT" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

echo '=== Enable 2FA ==='
curl -s -X PUT http://localhost/api/v1/system/2fa/config -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"enabled":true,"issuer":"ChatAgent"}'
echo

echo '=== Login now requires 2FA setup ==='
curl -s -X POST http://localhost/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"agent1@chatme365.online","password":"Ag&1234Bbo"}' | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; print('require_2fa_setup:', 'require_2fa_setup' in d)"

echo '=== Disable 2FA ==='
curl -s -X PUT http://localhost/api/v1/system/2fa/config -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"enabled":false,"issuer":"ChatAgent"}'
echo

echo '=== Login normal again ==='
curl -s -X POST http://localhost/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"agent1@chatme365.online","password":"Ag&1234Bbo"}' | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; print('has_token:', 'token' in d)"
