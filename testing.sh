#!/bin/bash
echo "-------------------------------------------"
echo "  HARDENED DEPLOYMENT VERIFICATION TESTING "
echo "-------------------------------------------"

echo ""
echo "[Test 1] Database connectivity -"
curl -s http://localhost:8081/api/dbstatus | grep -q '"db":"ok"' && echo "PASS: Database reachable" || echo "FAIL: Database unreachable"

echo ""
echo "[Test 2] Todos endpoint functioning -"
curl -s http://localhost:8081/api/todos | grep -q '\[' && echo "PASS: Todos endpoint responding" || echo "FAIL: Todos endpoint not responding"

echo ""
echo "[Test 3] Frontend serving correctly -"
curl -s http://localhost:8080 | grep -q "Todo List App" && echo "PASS: Frontend serving correctly" || echo "FAIL: Frontend not serving"

echo ""
echo "[Test 4] TC-03 - Database port NOT exposed to host -"
curl -s --connect-timeout 2 http://localhost:5432 > /dev/null 2>&1
if [ $? -ne 0 ]; then echo "PASS: Port 5432 not accessible from host"; else echo "FAIL: Port 5432 is exposed!"; fi

echo ""
echo "[Test 5] TC-05 - Containers running as non-root..."
SERVER_USER=$(docker exec exercise-backend-hardened ps -eo user,comm | awk '$2 == "server" {print $1}') 
FRONTEND_USER=$(docker exec exercise-frontend-hardened whoami)
echo "Backend /server process owner: $SERVER_USER"
echo "Frontend container user: $FRONTEND_USER"
if [ "$SERVER_USER" == "appuser" ] && [ "$FRONTEND_USER" == "www-data" ]; then
    echo "PASS: Both services running as non-root"
else
    echo "FAIL: Root user detected"
fi

echo ""
echo "[Test 6] TC-07 - Network isolation (frontend should not reach the database) -"
docker inspect exercise-frontend-hardened | grep -q "hardened_backend-network" && echo "FAIL: Frontend has backend network access!" || echo "PASS: Frontend isolated from backend-network"

echo ""
echo "[Test 7] TC-09 - Resource limits applied -"
docker inspect exercise-backend-hardened --format='{{.HostConfig.Memory}}' | grep -qv "^0$" && echo "PASS: Backend memory limit set" || echo "FAIL: No memory limit"

echo ""
echo "[Test 8] TC-10 - Health checks reporting healthy -"
docker compose ps | grep -q "healthy" && echo "PASS: Containers report healthy" || echo "FAIL: Containers not healthy"

echo ""
echo "[Test 9] Security headers present (bonus finding) -"
curl -sI http://localhost:8080 | grep -q "X-Frame-Options" && echo "PASS: Security headers present" || echo "FAIL: Security headers missing"

echo ""
echo "------------------------------------"
echo "  TESTING  COMPLETE"
echo "------------------------------------"
