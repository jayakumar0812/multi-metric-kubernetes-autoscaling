#!/bin/bash
echo "🔍 EMERGENCY DIAGNOSTIC: What's causing 24,019 complexity?"

echo "1. Checking GraphQL query cache:"
kubectl port-forward svc/graphql-server 4001:4001 &
PF1=$!
sleep 3

echo "Recent cached queries:"
curl -s http://localhost:4001/complexity-stats | jq '{
  cacheSize: .cacheSize,
  highest_complexity: [.cachedQueries | sort_by(.complexity) | reverse | .[0:3]]
}'

kill $PF1

echo -e "\n2. Checking hybrid metrics fusion stats:"
kubectl port-forward svc/hybrid-metrics-server 3001:3001 &
PF2=$!
sleep 3

curl -s http://localhost:3001/fusion-stats | jq '{
  average_complexity: .average_complexity,
  total_decisions: .total_decisions,
  recent_decision: .last_decision
}'

kill $PF2

echo -e "\n3. Checking custom metrics API:"
kubectl port-forward svc/custom-metrics-api 8443:8443 &
PF3=$!
sleep 3

curl -sk https://localhost:8443/status | jq '{
  current_complexity: .current_complexity,
  complexity_source: .complexity_source
}'

kill $PF3
