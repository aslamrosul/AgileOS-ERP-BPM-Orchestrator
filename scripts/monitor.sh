#!/bin/bash

# AgileOS Container Monitoring Script
# Monitors Docker container resource usage in real-time

echo "========================================="
echo "AgileOS Container Resource Monitor"
echo "========================================="
echo ""
echo "Press Ctrl+C to stop monitoring"
echo ""

# Check if docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running"
    exit 1
fi

# Monitor containers
while true; do
    clear
    echo "========================================="
    echo "AgileOS Container Resource Monitor"
    echo "Time: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "========================================="
    echo ""
    
    # Show container stats
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}" \
        agileos-db agileos-nats 2>/dev/null
    
    echo ""
    echo "========================================="
    echo "Container Health Status"
    echo "========================================="
    
    # Check health status
    for container in agileos-db agileos-nats; do
        if docker ps --filter "name=$container" --format "{{.Names}}" | grep -q "$container"; then
            health=$(docker inspect --format='{{.State.Health.Status}}' $container 2>/dev/null || echo "no healthcheck")
            status=$(docker inspect --format='{{.State.Status}}' $container)
            echo "$container: $status (health: $health)"
        else
            echo "$container: NOT RUNNING"
        fi
    done
    
    echo ""
    echo "Refreshing in 5 seconds... (Ctrl+C to stop)"
    sleep 5
done
