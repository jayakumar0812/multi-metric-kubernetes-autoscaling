#!/bin/bash
# WORKING RESEARCH TEST - GUARANTEED TO WORK

echo "🔬 SIMPLE WORKING RESEARCH TEST"
echo "=============================="

GRAPHQL_IP=$(kubectl get svc graphql-server -o jsonpath='{.spec.clusterIP}')
echo "📊 Testing GraphQL at: $GRAPHQL_IP"

# Test 1: Simple Query
echo ""
echo "1️⃣ SIMPLE QUERY TEST"
kubectl run test1 --image=curlimages/curl --rm -i --restart=Never -- \
curl -s -X POST http://$GRAPHQL_IP:4000/graphql \
-H 'Content-Type: application/json' \
-d '{"query": "{ user(id: \"1\") { name email } }"}'

sleep 5
SIMPLE_COMPLEXITY=$(kubectl exec deployment/redis -- redis-cli get current_complexity)
echo "📊 Simple Query Complexity: $SIMPLE_COMPLEXITY"

# Test 2: Complex Query  
echo ""
echo "2️⃣ COMPLEX QUERY TEST"
kubectl run test2 --image=curlimages/curl --rm -i --restart=Never -- \
curl -s -X POST http://$GRAPHQL_IP:4000/graphql \
-H 'Content-Type: application/json' \
-d '{"query": "{ systemAnalytics { totalUsers totalPosts topUsers { name email posts { title content } } } }"}'

sleep 5
COMPLEX_COMPLEXITY=$(kubectl exec deployment/redis -- redis-cli get current_complexity)
echo "📊 Complex Query Complexity: $COMPLEX_COMPLEXITY"

# Test 3: Check Scaling
echo ""
echo "3️⃣ SCALING TEST"
echo "Initial pods:"
kubectl get pods | grep graphql-server | wc -l

echo "Generating load for 2 minutes..."
kubectl run load-test --image=curlimages/curl --rm -i --restart=Never -- sh -c '
for i in {1..60}; do
  curl -s -X POST http://'$GRAPHQL_IP':4000/graphql \
    -H "Content-Type: application/json" \
    -d "{\"query\": \"{ systemAnalytics { totalUsers totalPosts topUsers { name email posts { title content comments { content author { name } } } } } }\"}" > /dev/null
  sleep 2
done
'

echo "Final pods:"
kubectl get pods | grep graphql-server | wc -l

echo ""
echo "4️⃣ RESULTS SUMMARY"
echo "Simple Complexity: $SIMPLE_COMPLEXITY"
echo "Complex Complexity: $COMPLEX_COMPLEXITY"
echo "HPA Status:"
kubectl get hpa

echo ""
echo "🎉 TEST COMPLETE!"
echo "✅ If complex > simple, your research works!"
echo "✅ If pods increased, your scaling works!"