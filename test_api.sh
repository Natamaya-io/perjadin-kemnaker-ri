#!/bin/bash
TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login -H "Content-Type: application/json" -d '{"email":"superadmin","password":"12345678"}' | jq -r .token)
curl -s "http://localhost:8081/api/v1/records/paginated?limit=50&sort_by=spj-desc" -H "Authorization: Bearer $TOKEN" > result.json
ls -lh result.json
