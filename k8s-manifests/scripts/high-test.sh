echo "Testing traditional system with high complexity queries..."

# Record baseline
echo "=== BASELINE BEFORE HIGH COMPLEXITY TEST ==="
kubectl get hpa traditional-cpu-hpa
echo "Pods: $(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)"

# Start monitoring again
watch -n 5 'echo "=== $(date) ==="; kubectl get hpa traditional-cpu-hpa; echo "Pods: $(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)"; kubectl top pods -l app=traditional-graphql-server --no-headers' &

# Test high complexity queries
kubectl exec $LOAD_POD -- sh -c "
echo 'Testing high complexity queries on traditional system...'
for i in \$(seq 1 15); do
  curl -s -X POST http://traditional-graphql-server:4000/graphql \
    -H 'Content-Type: application/json' \
    -d '{\"query\":\"query HighComplexity { systemAnalytics { totalUsers totalPosts topUsers { name posts { title comments { content author { name } } } } } }\"}' &
done
wait
echo 'High complexity test completed'
"

# Wait for traditional scaling (it will be slow - 5+ minutes)
echo "Waiting for traditional system to respond (this will be slow)..."
sleep 360  # 6 minutes

# Record results
echo "=== RESULTS AFTER HIGH COMPLEXITY TEST ==="
kubectl get hpa traditional-cpu-hpa
echo "Final pods: $(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)"
kubectl describe hpa traditional-cpu-hpa | grep Events -A 10

# Stop monitoring
pkill -f watch