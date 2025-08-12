#!/bin/bash
echo "🔍 DEBUGGING: What's really happening with queries..."

echo "=== Initial State ==="
kubectl port-forward svc/graphql-server 4001:4001 &
PF_PID=$!
sleep 2
echo "Cache size before:" 
curl -s http://localhost:4001/complexity-stats | jq '.cacheSize'
kill $PF_PID

echo -e "\n=== Sending Truly Different Queries ==="

queries=(
  '{"query":"query { systemAnalytics { totalUsers } }"}'
  '{"query":"query { allUsersWithPostsAndComments { name } }"}'
  '{"query":"query { userAnalytics(userId: \"1\") { totalPosts } }"}'
  '{"query":"query { users(limit: 5) { name posts { title } } }"}'
  '{"query":"query { posts(limit: 3) { title author { name } } }"}'
)

for i in "${!queries[@]}"; do
    query_num=$((i+1))
    echo "📤 Sending query $query_num: ${queries[i]:0:50}..."
    
    curl --request POST \
      --header 'content-type: application/json' \
      --url http://localhost:4000/graphql \
      --data "${queries[i]}" \
      --silent > /dev/null
    
    sleep 3
    
    # Check cache size after each query
    kubectl port-forward svc/graphql-server 4001:4001 &
    PF_PID=$!
    sleep 2
    
    echo "📊 Cache size after query $query_num:"
    curl -s http://localhost:4001/complexity-stats | jq '.cacheSize'
    
    echo "📊 Hybrid metrics complexity:"
    kill $PF_PID
    kubectl port-forward svc/hybrid-metrics-server 3001:3001 &
    PF_PID=$!
    sleep 2
    curl -s http://localhost:3001/fusion-stats | jq '.average_complexity'
    kill $PF_PID
    
    echo "---"
done

echo "🎯 Final check - HPA status:"
kubectl get hpa graphql-complexity-hpa
