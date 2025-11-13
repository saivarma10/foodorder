# API Curl Commands

## List API
```bash
curl -X GET http://localhost:8081/
```

## Create Table API
```bash
curl -X POST http://localhost:8081/createtable
```

## Insert Data API
```bash
curl -X POST http://localhost:8081/insert -H "Content-Type: application/json" -d '{"key": "value"}'
```

## Create Restaurant API
```bash
curl -X POST http://localhost:8081/createrestaurant -H "Content-Type: application/json" -d '{"name": "Restaurant Name", "location": "Location"}'
```

## Add Menu Item API
```bash
curl -X POST http://localhost:8081/addmenuitem \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_name": "Restaurant Name2",
    "item_name": "Item Name",
    "price": 10.99
  }'
```

## Get Menu by Restaurant API
```bash
curl -X POST http://localhost:8081/getmenubyrestaurant \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_name": "Restaurant Name2"
  }'
```

## Create Delivery API
```bash
 curl -X POST http://localhost:8081/createdelivery   -H "Content-Type: application/json"   -d '{
    "restaurant_name": "Pizza Palace",
    "restaurant_id": 1,
    "order_id": 1001,
    "delivery_name": "John Driver",
    "delivery_phone": "+1234567890",
    "delivery_location": "123 Main St, Apt 4B",
    "customer_name": "Jane Customer",
    "customer_phone": "+0987654321",
    "total_amount": 25.99
  }'
```