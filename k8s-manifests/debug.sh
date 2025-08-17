#!/bin/bash
# ============================================
# APPLICATION DEBUGGING AND FIX SCRIPT
# ============================================

echo "🔍 DEBUGGING APPLICATION ISSUES..."

# 1. CHECK CURRENT POD STATUS
echo "📋 Step 1: Checking Pod Status..."
kubectl get pods -o wide
echo ""

# 2. CHECK WHAT PORTS YOUR APPLICATIONS ARE ACTUALLY LISTENING ON
echo "📋 Step 2: Checking Application Ports..."

echo "🔍 GraphQL Server - checking ports:"
kubectl exec deployment/graphql-server -- netstat -tlnp 2>/dev/null || echo "❌ netstat not available"
kubectl exec deployment/graphql-server -- ss -tlnp 2>/dev/null || echo "❌ ss not available"

echo "🔍 Hybrid Metrics Server - checking ports:"
kubectl exec deployment/hybrid-metrics-server -- netstat -tlnp 2>/dev/null || echo "❌ netstat not available"
kubectl exec deployment/hybrid-metrics-server -- ss -tlnp 2>/dev/null || echo "❌ ss not available"

# 3. TEST ENDPOINTS DIRECTLY
echo "📋 Step 3: Testing Endpoints Directly..."

echo "🔍 Testing GraphQL server port 4001 (should be Prometheus format):"
kubectl exec deployment/graphql-server -- curl -s -I http://localhost:4001/metrics || echo "❌ GraphQL 4001 failed"
kubectl exec deployment/graphql-server -- curl -s http://localhost:4001/metrics | head -5 || echo "❌ GraphQL 4001 content failed"

echo "🔍 Testing GraphQL server port 4000 (GraphQL endpoint):"
kubectl exec deployment/graphql-server -- curl -s -I http://localhost:4000/ || echo "❌ GraphQL 4000 failed"

echo "🔍 Testing Hybrid Metrics server port 8080 (should be Prometheus format):"
kubectl exec deployment/hybrid-metrics-server -- curl -s -I http://localhost:8080/metrics || echo "❌ Hybrid 8080 failed"
kubectl exec deployment/hybrid-metrics-server -- curl -s http://localhost:8080/metrics | head -5 || echo "❌ Hybrid 8080 content failed"

echo "🔍 Testing Hybrid Metrics server port 3001:"
kubectl exec deployment/hybrid-metrics-server -- curl -s -I http://localhost:3001/ || echo "❌ Hybrid 3001 failed"
kubectl exec deployment/hybrid-metrics-server -- curl -s http://localhost:3001/ | head -5 || echo "❌ Hybrid 3001 content failed"

# 4. CHECK APPLICATION LOGS
echo "📋 Step 4: Checking Application Logs..."
echo "🔍 GraphQL Server logs:"
kubectl logs deployment/graphql-server --tail=10

echo "🔍 Hybrid Metrics Server logs:"
kubectl logs deployment/hybrid-metrics-server --tail=10

echo "🔍 Custom Metrics API logs:"
kubectl logs deployment/custom-metrics-api --tail=10

# 5. CHECK SERVICE CONNECTIVITY
echo "📋 Step 5: Testing Service Connectivity..."

echo "🔍 Testing from DataDog agent to services:"
DATADOG_POD=$(kubectl get pods -l app=datadog-agent -o jsonpath='{.items[0].metadata.name}')

if [ ! -z "$DATADOG_POD" ]; then
    echo "Testing connectivity from DataDog agent ($DATADOG_POD):"
    kubectl exec $DATADOG_POD -- curl -s -I http://graphql-server:4001/metrics || echo "❌ DataDog can't reach GraphQL"
    kubectl exec $DATADOG_POD -- curl -s -I http://hybrid-metrics-server:8080/metrics || echo "❌ DataDog can't reach Hybrid 8080"
    kubectl exec $DATADOG_POD -- curl -s -I http://hybrid-metrics-server:3001/ || echo "❌ DataDog can't reach Hybrid 3001"
fi

# 6. CHECK ENVIRONMENT VARIABLES
echo "📋 Step 6: Checking Environment Variables..."
echo "🔍 GraphQL Server environment:"
kubectl exec deployment/graphql-server -- printenv | grep -E "(PORT|GRAPHQL|REST|DD_)" || echo "No relevant env vars"

echo "🔍 Hybrid Metrics Server environment:"
kubectl exec deployment/hybrid-metrics-server -- printenv | grep -E "(PORT|GRAPHQL|HYBRID|DD_)" || echo "No relevant env vars"

echo ""
echo "🔧 POTENTIAL FIXES:"
echo "========================"

echo "🔧 Fix 1: If GraphQL /metrics returns JSON instead of Prometheus format:"
echo "   - Your application needs to expose metrics in Prometheus format"
echo "   - Expected: 'Content-Type: text/plain; version=0.0.4; charset=utf-8'"
echo "   - Current: 'application/json'"
echo ""

echo "🔧 Fix 2: If Hybrid Metrics port 8080 is not responding:"
echo "   - Check if your app is actually listening on port 8080"
echo "   - Verify the container port configuration"
echo "   - Check application startup logs for errors"
echo ""

echo "🔧 Fix 3: Update DataDog annotations if ports are different:"
echo "   - If metrics are on different ports, update the annotations in your deployment"
echo ""

echo "📋 NEXT STEPS:"
echo "1. Review the test results above"
echo "2. Fix application code to expose proper Prometheus metrics"
echo "3. Ensure correct port configuration"
echo "4. Update DataDog annotations if needed"
echo "5. Restart applications after fixes"

echo ""
echo "🔍 Run this to see what DataDog is actually trying to scrape:"
echo "kubectl logs -l app=datadog-agent | grep -E '(prometheus|metrics|curl)' | tail -20"
