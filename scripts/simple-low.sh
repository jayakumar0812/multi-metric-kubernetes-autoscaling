# Send simple queries with fixed names
for i in {1..20}; do
  curl -X POST http://localhost:4000/graphql \
    -H "Content-Type: application/json" \
    -d '{"query":"query { user(id: \"1\") { name } }"}' &
done
