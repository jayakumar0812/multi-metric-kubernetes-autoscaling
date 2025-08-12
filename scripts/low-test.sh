#!/bin/bash
echo "📉 SCALE-DOWN TEST: Sending simple queries..."

# Send 30 simple queries to lower complexity
for i in {1..30}; do
    curl --request POST \
      --header 'content-type: application/json' \
      --url http://localhost:4000/graphql \
      --data '{"query":"query Simple'$i' { __typename }"}' \
      --silent > /dev/null
    
    if [ $((i % 10)) -eq 0 ]; then
        echo "📤 Sent $i simple queries..."
    fi
    sleep 0.5
done

echo "✅ Scale-down queries sent!"
echo "📊 Complexity should drop below 30k → Should scale down from 8 pods"
