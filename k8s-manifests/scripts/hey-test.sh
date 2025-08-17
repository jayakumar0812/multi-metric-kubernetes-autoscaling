#!/bin/bash
# ============================================
# ADVANCED HEY STRESS TESTING SCRIPT
# Multiple complexity levels running simultaneously
# ============================================

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Starting Advanced GraphQL Stress Test${NC}"

# Start port forwarding
echo -e "${YELLOW}📡 Setting up port forwarding...${NC}"
kubectl port-forward svc/traditional-graphql-server 4000:4000 &
PORT_FORWARD_PID=$!
sleep 5

# Test connectivity
echo -e "${YELLOW}🔍 Testing connectivity...${NC}"
curl -s -X POST http://localhost:4000/graphql \
  -H 'Content-Type: application/json' \
  -d '{"query":"query { user(id: \"1\") { name } }"}' > /dev/null

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Connection successful${NC}"
else
    echo -e "${RED}❌ Connection failed${NC}"
    kill $PORT_FORWARD_PID 2>/dev/null
    exit 1
fi

# Check baseline
echo -e "${BLUE}📊 Baseline Status:${NC}"
kubectl get hpa graphql-complexity-hpa --no-headers | awk '{print "HPA:", $3, "Pods:", $7}'
kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1/namespaces/default/graphql_complexity_score" 2>/dev/null | jq -r '.items[0].value // "No complexity data"' | xargs echo "Complexity:"

echo -e "${YELLOW}⚡ Starting multi-level stress test...${NC}"

# ============================================
# STRESS TEST 1: Simple Query Flood
# ============================================
echo -e "${BLUE}🧪 Test 1: Simple Query Flood (Background Load)${NC}"
hey -n 2000 -c 20 -m POST \
  -H "Content-Type: application/json" \
  -d '{"query":"query Simple { user(id: \"1\") { name email } }"}' \
  http://localhost:4000/graphql > simple_results.txt 2>&1 &
SIMPLE_PID=$!

sleep 2

# ============================================
# STRESS TEST 2: Medium Complexity Burst
# ============================================
echo -e "${BLUE}🧪 Test 2: Medium Complexity Queries${NC}"
hey -n 800 -c 40 -m POST \
  -H "Content-Type: application/json" \
  -d '{"query":"query Medium { users { name email posts { title content comments { content author { name } } } } }"}' \
  http://localhost:4000/graphql > medium_results.txt 2>&1 &
MEDIUM_PID=$!

sleep 3

# ============================================
# STRESS TEST 3: High Complexity Analytics
# ============================================
echo -e "${BLUE}🧪 Test 3: High Complexity Analytics${NC}"
hey -n 400 -c 30 -m POST \
  -H "Content-Type: application/json" \
  -d '{"query":"query HighComplexity { systemAnalytics { totalUsers totalPosts averageComplexity topUsers { id name email posts { id title content tags comments { id content author { name email posts { id title } } } } } recentPosts { id title content author { name posts { title } } comments { content author { name } } } } }"}' \
  http://localhost:4000/graphql > high_results.txt 2>&1 &
HIGH_PID=$!

sleep 3

# ============================================
# STRESS TEST 4: Maximum Complexity Bomb
# ============================================
echo -e "${BLUE}🧪 Test 4: Maximum Complexity Queries${NC}"
hey -n 200 -c 50 -m POST \
  -H "Content-Type: application/json" \
  -d '{"query":"query MaxComplexity { systemAnalytics { totalUsers totalPosts averageComplexity topUsers { id name email createdAt posts { id title content tags createdAt comments { id content createdAt author { id name email createdAt posts { id title content tags createdAt comments { id content createdAt author { id name email createdAt posts { id title content comments { content author { name posts { title } } } } } } } } } } } recentPosts { id title content createdAt author { id name email createdAt posts { id title content tags createdAt comments { content author { name posts { title content } } } } } comments { id content createdAt author { id name email createdAt posts { id title content tags createdAt } } } } } }"}' \
  http://localhost:4000/graphql > max_results.txt 2>&1 &
MAX_PID=$!

sleep 3

# ============================================
# STRESS TEST 5: User Analytics Deep Dive
# ============================================
echo -e "${BLUE}🧪 Test 5: User Analytics Deep Dive${NC}"
hey -n 300 -c 25 -m POST \
  -H "Content-Type: application/json" \
  -d '{"query":"query UserAnalyticsDeep { userAnalytics(userId: \"1\") { user { id name email posts { id title content tags comments { id content author { name email posts { id title content comments { content author { name posts { title } } } } } } } } totalPosts totalComments averagePostLength mostPopularPost { id title content comments { id content author { name email posts { title } } } } engagementScore } }"}' \
  http://localhost:4000/graphql > analytics_results.txt 2>&1 &
ANALYTICS_PID=$!

# ============================================
# REAL-TIME MONITORING
# ============================================
echo -e "${YELLOW}📊 Monitoring system under load...${NC}"

# Monitor for 2 minutes
for i in {1..24}; do
    sleep 5
    
    # Get current metrics
    timestamp=$(date '+%H:%M:%S')
    cpu_usage=$(kubectl top pods -l app=graphql-server --no-headers 2>/dev/null | awk '{sum+=$2} END {print sum "m"}')
    pod_count=$(kubectl get pods -l app=graphql-server --no-headers | wc -l)
    hpa_status=$(kubectl get hpa graphql-complexity-hpa --no-headers 2>/dev/null | awk '{print $3}')
    complexity=$(kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1/namespaces/default/graphql_complexity_score" 2>/dev/null | jq -r '.items[0].value // "N/A"')
    
    # Color code the output based on scaling status
    if [ $pod_count -gt 2 ]; then
        echo -e "${GREEN}$timestamp - CPU: $cpu_usage, Pods: $pod_count, HPA: $hpa_status, Complexity: $complexity 🎯 SCALING!${NC}"
    elif [[ "$complexity" != "N/A" && "$complexity" -gt 25000 ]]; then
        echo -e "${YELLOW}$timestamp - CPU: $cpu_usage, Pods: $pod_count, HPA: $hpa_status, Complexity: $complexity ⚡ HIGH COMPLEXITY${NC}"
    else
        echo -e "$timestamp - CPU: $cpu_usage, Pods: $pod_count, HPA: $hpa_status, Complexity: $complexity"
    fi
    
    # Check if any major scaling occurred
    if [ $pod_count -gt 4 ]; then
        echo -e "${GREEN}🎉 MAJOR SCALING DETECTED! System scaled to $pod_count pods!${NC}"
        break
    fi
done

# ============================================
# WAIT FOR ALL TESTS TO COMPLETE
# ============================================
echo -e "${YELLOW}⏳ Waiting for all stress tests to complete...${NC}"

wait $SIMPLE_PID 2>/dev/null
echo -e "${GREEN}✅ Simple query test completed${NC}"

wait $MEDIUM_PID 2>/dev/null  
echo -e "${GREEN}✅ Medium complexity test completed${NC}"

wait $HIGH_PID 2>/dev/null
echo -e "${GREEN}✅ High complexity test completed${NC}"

wait $MAX_PID 2>/dev/null
echo -e "${GREEN}✅ Maximum complexity test completed${NC}"

wait $ANALYTICS_PID 2>/dev/null
echo -e "${GREEN}✅ Analytics deep dive test completed${NC}"

# ============================================
# FINAL RESULTS
# ============================================
echo -e "${BLUE}📊 FINAL RESULTS:${NC}"

# HPA Status
kubectl get hpa graphql-complexity-hpa

# Pod Status  
echo -e "\n${BLUE}Pod Status:${NC}"
kubectl get pods -l app=graphql-server

# Resource Usage
echo -e "\n${BLUE}Resource Usage:${NC}"
kubectl top pods -l app=graphql-server

# Complexity Score
echo -e "\n${BLUE}Final Complexity Score:${NC}"
kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1/namespaces/default/graphql_complexity_score" 2>/dev/null | jq . || echo "Complexity metrics not available"

# HPA Events
echo -e "\n${BLUE}HPA Events:${NC}"
kubectl describe hpa graphql-complexity-hpa | grep Events -A 10

# ============================================
# PERFORMANCE SUMMARY
# ============================================
echo -e "\n${BLUE}📈 PERFORMANCE SUMMARY:${NC}"

if [ -f simple_results.txt ]; then
    simple_rps=$(grep "Requests/sec:" simple_results.txt | awk '{print $2}')
    echo "Simple Queries: $simple_rps req/sec"
fi

if [ -f medium_results.txt ]; then
    medium_rps=$(grep "Requests/sec:" medium_results.txt | awk '{print $2}')
    echo "Medium Queries: $medium_rps req/sec"
fi

if [ -f high_results.txt ]; then
    high_rps=$(grep "Requests/sec:" high_results.txt | awk '{print $2}')
    echo "High Complexity: $high_rps req/sec"
fi

if [ -f max_results.txt ]; then
    max_rps=$(grep "Requests/sec:" max_results.txt | awk '{print $2}')
    echo "Max Complexity: $max_rps req/sec"
fi

if [ -f analytics_results.txt ]; then
    analytics_rps=$(grep "Requests/sec:" analytics_results.txt | awk '{print $2}')
    echo "Analytics: $analytics_rps req/sec"
fi

# ============================================
# CLEANUP
# ============================================
echo -e "\n${YELLOW}🧹 Cleaning up...${NC}"
kill $PORT_FORWARD_PID 2>/dev/null
rm -f simple_results.txt medium_results.txt high_results.txt max_results.txt analytics_results.txt

# Final pod count
final_pods=$(kubectl get pods -l app=graphql-server --no-headers | wc -l)
if [ $final_pods -gt 2 ]; then
    echo -e "${GREEN}🎯 SUCCESS! System scaled to $final_pods pods!${NC}"
    echo -e "${GREEN}Your intelligent auto-scaling system works!${NC}"
else
    echo -e "${YELLOW}⚠️  No scaling triggered. Try increasing query complexity or reducing resource limits.${NC}"
fi

echo -e "${BLUE}✅ Stress test completed!${NC}"
