#!/bin/bash
echo "🚀 Sending FRESH complex queries to spike complexity above 500..."

# Send multiple different complex queries to increase cache and complexity
queries=(
  '{"query":"query Fresh1 { systemAnalytics { totalUsers totalPosts topUsers { id name email posts { id title content comments { id content author { id name email } } } } } }"}'
  '{"query":"query Fresh2 { allUsersWithPostsAndComments { id name email posts { title content comments { content author { name } } } } }"}'
  '{"query":"query Fresh3 { userAnalytics(userId: \"1\") { user { name posts { title comments { author { name posts { title } } } } } totalPosts engagementScore } }"}'
  '{"query":"query Fresh4 { userAnalytics(userId: \"2\") { user { name posts { title comments { author { name posts { title } } } } } totalPosts engagementScore } }"}'
  '{"query":"query Fresh5 { systemAnalytics { totalUsers topUsers { name posts { title content tags comments { content author { name email } } } } recentPosts { title author { name } } } }"}'
)

for i in "${!queries[@]}"; do
    query_num=$((i+1))
    echo "📈 Sending fresh complex query #$query_num"
    
    curl --request POST \
      --header 'content-type: application/json' \
      --url http://localhost:4000/graphql \
      --data "${queries[i]}" \
      --silent > /dev/null
    
    echo "✅ Query #$query_num sent"
    sleep 3
done

echo "🎯 All fresh queries sent! Complexity should spike now..."
