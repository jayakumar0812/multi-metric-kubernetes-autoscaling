#!/bin/bash
# ============================================
# LOAD TEST USING PORT FORWARDING
# Uses localhost:4000 via kubectl port-forward
# ============================================

echo "🚀 LOAD TEST VIA PORT FORWARDING"
echo "================================"

GRAPHQL_URL="http://localhost:4000/graphql"
DATA_DIR="localhost_test_$(date +%Y%m%d_%H%M%S)"
mkdir -p $DATA_DIR

echo "📊 GraphQL URL: $GRAPHQL_URL"
echo "📁 Data Directory: $DATA_DIR"
echo "🔧 Make sure port forwarding is running: kubectl port-forward svc/traditional-graphql-server 4000:4000"
echo ""

# Test connection first
echo "🔍 Testing connection..."
curl -s -X POST $GRAPHQL_URL \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __schema { types { name } } }"}' > /dev/null

if [ $? -eq 0 ]; then
    echo "✅ Connection successful!"
else
    echo "❌ Connection failed! Make sure port forwarding is running:"
    echo "   kubectl port-forward svc/traditional-graphql-server 4000:4000"
    exit 1
fi

# Reset to baseline
echo ""
echo "🔄 Resetting to baseline (2 pods)..."
kubectl scale deployment traditional-graphql-server --replicas=2
sleep 60

# Initialize CSV
echo "Load_Level,Complexity_Score,Pod_Count_Before,Pod_Count_After,Wait_Time,Scaling_Result" > $DATA_DIR/localhost_results.csv

# ============================================
# TEST 1: LOW LOAD
# ============================================
echo ""
echo "🧪 TEST 1: LOW Load (Light Traffic)"
echo "==================================="

pods_before=$(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)
echo "📊 Pods before: $pods_before"

echo "🔹 Running light load test (50 requests, 5 concurrent)..."
hey -n 50 -c 5 -t 30 -m POST \
  -H "Content-Type: application/json" \
  -d '{"query":"{ user(id: \"1\") { name email } }"}' \
  $GRAPHQL_URL > $DATA_DIR/hey_low_output.txt 2>&1

echo "📊 Low load test results:"
grep -E "(Requests/sec|Average|Status code|Error)" $DATA_DIR/hey_low_output.txt | head -5

echo "⏳ Waiting for scaling response (3 minutes)..."
for i in {1..6}; do
    sleep 30
    current_pods=$(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)
    echo "   ${i}*30s: $current_pods pods"
done

pods_after=$(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)
scaling_result=$([ $pods_after -gt $pods_before ] && echo "SCALED_UP" || echo "NO_SCALING")

echo "low,100,$pods_before,$pods_after,180,$scaling_result" >> $DATA_DIR/localhost_results.csv
echo "✅ LOW LOAD COMPLETE: $pods_before → $pods_after pods ($scaling_result)"

# Cooldown
echo "🔄 Cooldown (90 seconds)..."
sleep 90

# ============================================
# TEST 2: MEDIUM LOAD
# ============================================
echo ""
echo "🧪 TEST 2: MEDIUM Load (Moderate Traffic)"
echo "========================================"

pods_before=$(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)
echo "📊 Pods before: $pods_before"

echo "🔹 Running medium load test (200 requests, 10 concurrent)..."
hey -n 200 -c 10 -t 30 -m POST \
  -H "Content-Type: application/json" \
  -d '{"query":"{ users(limit: 10) { name email posts { title } } }"}' \
  $GRAPHQL_URL > $DATA_DIR/hey_medium_output.txt 2>&1

echo "📊 Medium load test results:"
grep -E "(Requests/sec|Average|Status code|Error)" $DATA_DIR/hey_medium_output.txt | head -5

echo "⏳ Waiting for scaling response (3 minutes)..."
for i in {1..6}; do
    sleep 30
    current_pods=$(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)
    echo "   ${i}*30s: $current_pods pods"
done

pods_after=$(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)
scaling_result=$([ $pods_after -gt $pods_before ] && echo "SCALED_UP" || echo "NO_SCALING")

echo "medium,400,$pods_before,$pods_after,180,$scaling_result" >> $DATA_DIR/localhost_results.csv
echo "✅ MEDIUM LOAD COMPLETE: $pods_before → $pods_after pods ($scaling_result)"

# Cooldown
echo "🔄 Cooldown (90 seconds)..."
sleep 90

# ============================================
# TEST 3: HIGH LOAD
# ============================================
echo ""
echo "🧪 TEST 3: HIGH Load (Heavy Traffic)"
echo "==================================="

pods_before=$(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)
echo "📊 Pods before: $pods_before"

echo "🔹 Running high load test (500 requests, 25 concurrent)..."
hey -n 500 -c 25 -t 30 -m POST \
  -H "Content-Type: application/json" \
  -d '{"query":"query HighUserData { userWithAllData(id: \"1\") { id name email posts { id title content comments { id content author { name } } } } }"}' \
  $GRAPHQL_URL > $DATA_DIR/hey_high_output.txt 2>&1

echo "📊 High load test results:"
grep -E "(Requests/sec|Average|Status code|Error)" $DATA_DIR/hey_high_output.txt | head -5

echo "⏳ Waiting for scaling response (4 minutes)..."
for i in {1..8}; do
    sleep 30
    current_pods=$(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)
    echo "   ${i}*30s: $current_pods pods"
    if [ $current_pods -gt $pods_before ]; then
        echo "   🎉 SCALING DETECTED!"
    fi
done

pods_after=$(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)
scaling_result=$([ $pods_after -gt $pods_before ] && echo "SCALED_UP" || echo "NO_SCALING")

echo "high,800,$pods_before,$pods_after,240,$scaling_result" >> $DATA_DIR/localhost_results.csv
echo "✅ HIGH LOAD COMPLETE: $pods_before → $pods_after pods ($scaling_result)"

# Cooldown
echo "🔄 Cooldown (2 minutes)..."
sleep 120

# ============================================
# TEST 4: VERY HIGH LOAD
# ============================================
echo ""
echo "🧪 TEST 4: VERY HIGH Load (Maximum Traffic)"
echo "=========================================="

pods_before=$(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)
echo "📊 Pods before: $pods_before"

echo "🔹 Running very high load test (1000 requests, 50 concurrent)..."
hey -n 1000 -c 50 -t 60 -m POST \
  -H "Content-Type: application/json" \
  -d '{"query":"query HighUserData { userWithAllData(id: \"1\") { id name email posts { id title content comments { id content author { name } } } } }"}' \
  $GRAPHQL_URL > $DATA_DIR/hey_very_high_output.txt 2>&1

echo "📊 Very high load test results:"
grep -E "(Requests/sec|Average|Status code|Error)" $DATA_DIR/hey_very_high_output.txt | head -5

echo "⏳ Waiting for scaling response (5 minutes)..."
for i in {1..10}; do
    sleep 30
    current_pods=$(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)
    echo "   ${i}*30s: $current_pods pods"
    if [ $current_pods -gt $pods_before ]; then
        echo "   🎉 SCALING DETECTED!"
    fi
done

pods_after=$(kubectl get pods -l app=traditional-graphql-server --no-headers | wc -l)
scaling_result=$([ $pods_after -gt $pods_before ] && echo "SCALED_UP" || echo "NO_SCALING")

echo "very_high,800,$pods_before,$pods_after,300,$scaling_result" >> $DATA_DIR/localhost_results.csv
echo "✅ VERY HIGH LOAD COMPLETE: $pods_before → $pods_after pods ($scaling_result)"

# ============================================
# ANALYZE RESULTS
# ============================================
echo ""
echo "📊 LOCALHOST LOAD TEST RESULTS"
echo "=============================="
echo ""

# Display CSV results
echo "📋 Final CSV Results:"
echo "────────────────────────────────────────────────────────────────────"
cat $DATA_DIR/localhost_results.csv
echo "────────────────────────────────────────────────────────────────────"
echo ""

# Performance summary
echo "📈 PERFORMANCE SUMMARY:"
echo "======================"

for level in low medium high very_high; do
    output_file="$DATA_DIR/hey_${level}_output.txt"
    if [ -f "$output_file" ]; then
        echo ""
        echo "🔹 $level load performance:"
        req_sec=$(grep "Requests/sec:" $output_file | awk '{print $2}' || echo "N/A")
        avg_time=$(grep "Average:" $output_file | awk '{print $2}' || echo "N/A")
        echo "   Requests/sec: $req_sec"
        echo "   Average latency: $avg_time"
        
        # Check for errors
        if grep -q "Error distribution:" $output_file; then
            error_count=$(grep -A 10 "Error distribution:" $output_file | grep -v "Error distribution:" | wc -l)
            echo "   Errors detected: $error_count types"
        else
            echo "   No errors detected"
        fi
    fi
done

# Scaling analysis
echo ""
echo "🎯 SCALING ANALYSIS:"
echo "==================="

total_scaled=0
while IFS=, read -r level complexity pods_before pods_after wait_time result; do
    if [ "$level" != "Load_Level" ]; then
        change=$((pods_after - pods_before))
        echo "📊 $level (complexity $complexity): $pods_before → $pods_after pods"
        if [ "$result" = "SCALED_UP" ]; then
            echo "   ✅ Scaled up by $change pods"
            total_scaled=$((total_scaled + 1))
        else
            echo "   ⚪ No scaling occurred"
        fi
    fi
done < $DATA_DIR/localhost_results.csv

echo ""
echo "🎯 SUMMARY:"
if [ $total_scaled -gt 0 ]; then
    echo "✅ SUCCESS: $total_scaled scaling events detected"
    echo "📊 Traditional CPU-based scaling is working with localhost testing"
else
    echo "⚠️  NO SCALING: Traditional system may need further tuning"
    echo "💡 Check HPA metrics and CPU thresholds"
fi

# Final system state
echo ""
echo "📋 FINAL SYSTEM STATE:"
echo "====================="
kubectl get hpa traditional-cpu-hpa
echo ""
kubectl get pods -l app=traditional-graphql-server
echo ""

# Check current CPU usage
echo "💻 Current resource usage:"
kubectl top pods -l app=traditional-graphql-server 2>/dev/null || echo "   Metrics server not available"

echo ""
echo "🎉 LOCALHOST LOAD TEST COMPLETE!"
echo "==============================="
echo "📊 CSV Results: $DATA_DIR/localhost_results.csv"
echo "📈 Performance Details: $DATA_DIR/hey_*_output.txt"
echo ""
echo "📋 Ready for graphing! Your CSV format:"
echo ""
cat $DATA_DIR/localhost_results.csv