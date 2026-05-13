#!/bin/bash
set -e

# Login with known credentials
LOGIN=$(curl -s http://localhost/api/v1/auth/login -H "Content-Type: application/json" -d '{"email":"root@chatme365.online","password":"Cha@365mexXy"}')
TOKEN=$(echo "$LOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('token',''))")
echo "Token: ${TOKEN:0:20}..."

# Create inbox
INBOX=$(curl -s http://localhost/api/v1/inboxes -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"name":"test-msg-'$(date +%s)'","inbox_type":"website","welcome_title":"Hello!","welcome_message":"Welcome to support!"}')
INBOX_ID=$(echo "$INBOX" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))")
echo "Inbox: $INBOX_ID"

# Widget auth
AUTH=$(curl -s http://localhost/api/v1/widget/auth -H "Content-Type: application/json" -d '{"inbox_id":"'$INBOX_ID'","fingerprint":"fp123"}')
PUBSUB=$(echo "$AUTH" | python3 -c "import sys,json; d=json.load(sys.stdin).get('data',{}); print(d.get('pubsub_token',''))")
CONTACT_ID=$(echo "$AUTH" | python3 -c "import sys,json; d=json.load(sys.stdin).get('data',{}); print(d.get('contact_id',''))")
echo "ContactID: $CONTACT_ID, Token: ${PUBSUB:0:20}..."

# Check contact name
CONTACT=$(curl -s http://localhost/api/v1/contacts/$CONTACT_ID -H "Authorization: Bearer $TOKEN")
CONTACT_NAME=$(echo "$CONTACT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('name',''))")
echo "Contact Name: $CONTACT_NAME"

# Send message
SEND=$(curl -s http://localhost/api/v1/widget/messages -H "Content-Type: application/json" -d '{"pubsub_token":"'$PUBSUB'","content":"Hello, I need help with my order"}')
CONV_ID=$(echo "$SEND" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('conversation_id',''))")
echo "Conversation: $CONV_ID"

# List conversations
echo "=== Conversation List ==="
curl -s http://localhost/api/v1/conversations -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
convs=d.get('data',{}).get('list',[])
print(f'Total: {d.get(\"data\",{}).get(\"total\",0)}')
for c in convs:
    print(f'  #{c.get(\"display_id\")} status={c.get(\"status\")} contact={c.get(\"contact_name\",\"?\")} subject={c.get(\"subject\",\"\")[:50]}')
"

echo ""
echo "=== Test completed ==="
