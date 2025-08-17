# Create realistic query mix that shows the problem
echo "=== REALISTIC QUERY PATTERN TEST ==="

# 90% simple queries (normal traffic)
echo "Phase 1: Normal traffic (90% simple queries)"
for i in {1..90}; do
  curl -s -X POST http://localhost:4000/graphql \
    -H 'Content-Type: application/json' \
    -d '{"query":"query { user(id: \"1\") { name } }"}' > /dev/null &
done

# 10% complex analytics queries (business intelligence)
echo "Phase 2: Analytics spike (10% complex queries)"
for i in {1..10}; do
  curl -s -X POST http://localhost:4000/graphql \
    -H 'Content-Type: application/json' \
    -d '{"query":"query Analytics { systemAnalytics { topUsers { name posts { title comments { content author { name posts { title } } } } } } }"}' > /dev/null &
done

wait
kubectl top pods -l app=traditional-graphql-server
kubectl get hpa traditional-cpu-hpa
