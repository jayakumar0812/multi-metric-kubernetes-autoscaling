# Get the load tester pod (with error checking)
LOAD_POD=$(kubectl get pods -l app=load-tester -o jsonpath='{.items[0].metadata.name}')

if [ -z "$LOAD_POD" ]; then
    echo "❌ Load tester pod not found. Checking deployment..."
    kubectl get deployment load-tester
    kubectl get pods -l app=load-tester
    exit 1
fi

echo "✅ Using load tester pod: $LOAD_POD"

# Start monitoring in background
echo "📊 Starting monitoring..."
(while true; do
    echo "=== $(date) ==="
    kubectl get hpa traditional-cpu-hpa
    echo "Pods: $(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)"
    kubectl top pods -l app=traditional-graphql-server --no-headers 2>/dev/null || echo "Metrics not ready"
    echo "---"
    sleep 10
done) &
MONITOR_PID=$!

# Test simple queries
echo "🧪 Testing simple queries..."
kubectl exec $LOAD_POD -- sh -c '
echo "Testing simple queries on traditional system..."
for i in $(seq 1 50); do
  curl -s -X POST http://traditional-graphql-server:4000/graphql \
    -H "Content-Type: application/json" \
    -d "{\"query\":\"query SimpleUser { user(id: \\\"1\\\") { id name email } }\"}" &
done
wait
echo "Simple query test completed"
'

# Wait and observe
echo "⏱️ Waiting 90 seconds to observe scaling behavior..."
sleep 90

# Stop monitoring
kill $MONITOR_PID 2>/dev/null

echo "📊 Final results after simple queries:"
kubectl get hpa traditional-cpu-hpa
echo "Final pods: $(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)"