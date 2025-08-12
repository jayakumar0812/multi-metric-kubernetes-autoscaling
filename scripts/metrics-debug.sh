#!/bin/bash
echo "📉 Sending simple queries to trigger scale-down..."

# Send 10 simple queries to lower average complexity
for i in {1..10}; do
    echo "📤 Sending simple query #$i"
    
    # Alternate between different simple queries
    if [ $((i % 3)) -eq 0 ]; then
        curl --request POST \
          --header 'content-type: application/json' \
          --url http://localhost:4000/graphql \
          --data '{"query":"query { user(id: \"'$i'\") { name } }"}' \
          --silent > /dev/null
    elif [ $((i % 3)) -eq 1 ]; then
        curl --request POST \
          --header 'content-type: application/json' \
          --url http://localhost:4000/graphql \
          --data '{"query":"query { post(id: \"'$i'\") { title } }"}' \
          --silent > /dev/null
    else
        curl --request POST \
          --header 'content-type: application/json' \
          --url http://localhost:4000/graphql \
          --data '{"query":"query { users(limit: 1) { name } }"}' \
          --silent > /dev/null
    fi
    
    sleep 2
done

echo "✅ Simple queries sent! Checking complexity..."
./metrics-debug.sh
