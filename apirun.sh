#!/bin/bash

# Usage: ./run.sh <port> use node port of kubectl get svc go-service
# Example: ./run.sh 8081

PORT="$1"

if [ -z "$PORT" ]; then
  echo "Usage: $0 <port>"
  exit 1
fi

echo "Using Port: $PORT"
echo

echo "Creating tables..."
curl -s -X POST "http://localhost:${PORT}/createtable"
echo

echo "Creating restaurant..."
curl -s -X POST "http://localhost:${PORT}/createrestaurant" \
  -H "Content-Type: application/json" \
  -d '{"name": "wave restarant", "location": "space"}'
echo

echo "Adding menu item..."
curl -s -X POST "http://localhost:${PORT}/addmenuitem" \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_name": "wave restarant",
    "item_name": "ebpf",
    "price": 100000.99
  }'
echo

echo "Creating delivery..."
curl -s -X POST "http://localhost:${PORT}/createdelivery" \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_name": "wave restarant",
    "restaurant_id": 2,
    "order_id": 1003,
    "delivery_name": "John Driver",
    "delivery_phone": "+1234567890",
    "delivery_location": "123 Main St, Apt 4B",
    "customer_name": "Jane Customer",
    "customer_phone": "+0987654321",
    "total_amount": 25.99
  }'
echo

echo "Done."