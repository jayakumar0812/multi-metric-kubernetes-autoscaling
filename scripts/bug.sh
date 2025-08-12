#!/bin/bash
echo "🔍 DEBUGGING: Exact values in the pipeline"

echo "1. What Custom Metrics API is sending to Kubernetes:"
kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1/namespaces/default/graphql_complexity_score" | jq '{
  raw_value: .items[0].value,
  as_number: (.items[0].value | tonumber),
  expected_if_235: "23500 (235 x 100)",
  actual_multiplier: ((.items[0].value | tonumber) / 23500)
}'

echo -e "\n2. Current fusion stats from hybrid server:"
kubectl port-forward svc/hybrid-metrics-server 3001:3001 &
PF_PID=$!
sleep 2
current_complexity=$(curl -s http://localhost:3001/fusion-stats | jq -r '.average_complexity')
echo "Current hybrid complexity: $current_complexity"
kill $PF_PID

echo -e "\n3. Custom Metrics API logs (what it's calculating):"
kubectl logs -l app=custom-metrics-api --tail=10 | grep -E "complexity|Returning"

echo -e "\n4. Expected calculation:"
expected=$(echo "$current_complexity * 100" | bc)
echo "Expected K8s value: $expected (${current_complexity} × 100)"
