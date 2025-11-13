package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	pb "food/proto"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"

	"database/sql"
	"time"
)

var DB *sql.DB

type FoodOrder struct {
	Item  string
	Price float64
}

type DeliveryPayload struct {
	RestaurantName   string  `json:"restaurant_name"`
	RestaurantID     int     `json:"restaurant_id"`
	OrderID          int     `json:"order_id"`
	DeliveryName     string  `json:"delivery_name"`
	DeliveryPhone    string  `json:"delivery_phone"`
	DeliveryLocation string  `json:"delivery_location"`
	CustomerName     string  `json:"customer_name"`
	CustomerPhone    string  `json:"customer_phone"`
	TotalAmount      float64 `json:"total_amount"`
}
type foodOrderServer struct {
	pb.UnimplementedFoodOrderServiceServer
}

func (s *foodOrderServer) GetMenu(ctx context.Context, req *pb.MenuRequest) (*pb.MenuResponse, error) {
	restaurantName := req.GetRestaurantName()

	conn, err := connectsql()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query("SELECT item_name, price FROM menu WHERE restaurant_name=$1", restaurantName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*pb.MenuItem
	for rows.Next() {
		var itemName string
		var price float64
		err := rows.Scan(&itemName, &price)
		if err != nil {
			return nil, err
		}
		items = append(items, &pb.MenuItem{
			ItemName: itemName,
			Price:    price,
		})
	}

	return &pb.MenuResponse{
		Items: items,
	}, nil
}

func startGRPCServer() {
	lis, err := net.Listen("tcp", ":51058")
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterFoodOrderServiceServer(s, &foodOrderServer{})

	log.Println("gRPC server on :51058")
	s.Serve(lis)
}

// postgres direct package
// func connect() (*pgx.Conn, error) {

// 	conn, err := pgx.Connect(context.Background(), "postgres://postgres:mysecretpassword@localhost:5432/postgres")
// 	if err != nil {
// 		return nil, err
// 	}
// 	return conn, nil
// }

// db sql for instrumentation testing
func connectkafka(topic string, partition int) (*kafka.Conn, error) {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}

	dialer := &kafka.Dialer{
		Timeout:  10 * time.Second,
		ClientID: "foodorder-client",
	}
	conn, err := dialer.DialLeader(context.Background(), "tcp", kafkaBroker, topic, partition)
	if err != nil {
		fmt.Println("failed to dial leader:", err)
		return nil, err
	}
	return conn, nil
}

// func connectkafka(topic string, partition int) (*kafka.Conn, error) {
// 	conn, err := kafka.DialLeader(context.Background(), "tcp",
// 		"localhost:9093", topic, partition)
// 	if err != nil {
// 		fmt.Println("failed to dial leader")
// 	}
// 	return conn, err
// } //end connect

func writeMessages(conn *kafka.Conn, msgs []string) {
	var err error
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	for _, msg := range msgs {
		_, err = conn.WriteMessages(
			kafka.Message{Value: []byte(msg)})
	}
	if err != nil {
		fmt.Println("failed to write messages:", err)
	}
} //end writeMessages

func readMessages(conn *kafka.Conn, minSize int, maxSize int) {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	batch := conn.ReadBatch(minSize, maxSize) //in bytes

	msg := make([]byte, 10e3) //set the max length of each message
	for {
		msgSize, err := batch.Read(msg)
		if err != nil {
			break
		}
		fmt.Println(string(msg[:msgSize]))
	}

	if err := batch.Close(); err != nil { //make sure to close the batch
		fmt.Println("failed to close batch:", err)
	}
} //end readMessages

func connectsql() (*sql.DB, error) {
	// Read from environment variables with defaults
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	portStr := os.Getenv("DB_PORT")
	port := 5432
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "mysecretpassword"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "postgres"
	}

	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	DB, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}
	return DB, nil
}
func queryData(conn *sql.DB) {
	rows, err := conn.Query("SELECT id, name FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User ID: %d, Name: %s\n", id, name)
	}
}

func createTableRestaurants(conn *sql.DB) {
	_, err := conn.Exec("CREATE TABLE if not exists restaurants (id SERIAL PRIMARY KEY, name TEXT UNIQUE NOT NULL, location TEXT NOT NULL)")
	if err != nil {
		log.Fatal(err)
	}
}
func createTableOrders(conn *sql.DB) {
	_, err := conn.Exec(`CREATE TABLE if not exists orders  (
		id SERIAL PRIMARY KEY,
		restaurant_id INTEGER NOT NULL,
		restaurant_name TEXT NOT NULL,
		customer_name TEXT NOT NULL,
		customer_phone TEXT,
		order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		total_amount NUMERIC(10, 2) NOT NULL,
		status TEXT DEFAULT 'pending',
		FOREIGN KEY (restaurant_name) REFERENCES restaurants(name)
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func createTableOrderItems(conn *sql.DB) {
	_, err := conn.Exec(`CREATE TABLE if not exists order_items (
		id SERIAL PRIMARY KEY,
		order_id INTEGER NOT NULL,
		item_name TEXT NOT NULL,
		quantity INTEGER NOT NULL,
		price NUMERIC(10, 2) NOT NULL,
		FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func createTableDeliveries(conn *sql.DB) {
	_, err := conn.Exec(`CREATE TABLE if not exists deliveries(
		id SERIAL PRIMARY KEY,
		order_id INTEGER UNIQUE NOT NULL,
		delivery_name TEXT NOT NULL,
		delivery_phone TEXT,
		delivery_location TEXT NOT NULL,
		delivery_status TEXT DEFAULT 'pending',
		assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		delivered_at TIMESTAMP,
		FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.Fatal(err)
	}
}
func createTableMenu(conn *sql.DB) {
	_, err := conn.Exec("CREATE TABLE if not exists menu (id SERIAL PRIMARY KEY, restaurant_name TEXT NOT NULL, item_name TEXT NOT NULL, price NUMERIC(10, 2) NOT NULL, FOREIGN KEY (restaurant_name) REFERENCES restaurants (name))")
	if err != nil {
		log.Fatal(err)
	}
}
func addNewRestaurant(conn *sql.DB, name string, location string) {
	_, err := conn.Exec("INSERT INTO restaurants (name, location) VALUES ($1, $2)", name, location)
	if err != nil {
		log.Fatal(err)
	}
}
func addMenuItem(conn *sql.DB, restaurantName string, itemName string, price float64) {
	_, err := conn.Exec("INSERT INTO menu (restaurant_name, item_name, price) VALUES ($1, $2, $3)", restaurantName, itemName, price)
	if err != nil {
		log.Fatal(err)
	}
}
func getMenuByRestaurant(conn *sql.DB, restaurantName string) {
	rows, err := conn.Query("SELECT item_name, price FROM menu WHERE restaurant_name=$1", restaurantName)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var itemName string
		var price float64
		err := rows.Scan(&itemName, &price)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Item: %s, Price: %.2f\n", itemName, price)
	}
}
func insertData(conn *sql.DB, id int, name string) {
	_, err := conn.Exec("INSERT INTO users (id, name) VALUES ($1, $2)", id, name)
	if err != nil {
		log.Fatal(err)
	}
}

func list(w http.ResponseWriter, r *http.Request) {
	order := FoodOrder{
		Item:  "Pizza",
		Price: 9.99,
	}

	conn, err := connectsql()
	if err != nil {
		log.Printf("Error connecting to database: %v\n", err)

		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	queryData(conn)
	orderJson, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "Error converting to JSON", http.StatusInternalServerError)
		return
	}
	w.Write(orderJson)
}
func createTableHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := connectsql()

	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	createTableRestaurants(conn)
	createTableMenu(conn)
	createTableOrders(conn)
	createTableOrderItems(conn)
	createTableDeliveries(conn)

	w.Write([]byte("Table created successfully"))
}

func createDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	var payload DeliveryPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Forward the request to Java API
	jsonData, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post("http://localhost:8085/delivery", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Error calling Java API: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Java API returned error", resp.StatusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Delivery created successfully"))
}

func addNewRestaurantHandler(w http.ResponseWriter, r *http.Request) {
	type InputPayload struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}
	var payload InputPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	name := payload.Name
	location := payload.Location

	conn, err := connectsql()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	addNewRestaurant(conn, name, location)
	w.Write([]byte("Restaurant added successfully"))
}
func addMenuItemHandler(w http.ResponseWriter, r *http.Request) {
	type InputPayload struct {
		RestaurantName string  `json:"restaurant_name"`
		ItemName       string  `json:"item_name"`
		Price          float64 `json:"price"`
	}
	var payload InputPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	restaurantName := payload.RestaurantName
	itemName := payload.ItemName
	price := payload.Price

	conn, err := connectsql()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	addMenuItem(conn, restaurantName, itemName, price)
	w.Write([]byte("Menu item added successfully"))
}

func getMenuByRestaurantHandler(w http.ResponseWriter, r *http.Request) {
	type InputPayload struct {
		RestaurantName string `json:"restaurant_name"`
	}
	var payload InputPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	restaurantName := payload.RestaurantName

	conn, err := connectsql()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	getMenuByRestaurant(conn, restaurantName)
	w.Write([]byte("Menu retrieved successfully"))
}

func insertDataHandler(w http.ResponseWriter, r *http.Request) {
	type InputPayload struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	var payload InputPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	id := payload.ID
	name := payload.Name

	conn, err := connectsql()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	insertData(conn, id, name)
	w.Write([]byte("Data inserted successfully"))
}

func main() {
	order := FoodOrder{
		Item:  "Pizza",
		Price: 9.99,
	}
	go startGRPCServer()

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}

	fmt.Println("Order Details:", order)
	topic := "wave"
	partition := 0
	c, err := kafka.Dial("tcp", kafkaBroker)
	if err != nil {
		fmt.Println("failed to connect to Kafka broker:", err)
		return
	}
	kt := kafka.TopicConfig{Topic: "wave", NumPartitions: 1, ReplicationFactor: 1}
	err = c.CreateTopics(kt)
	if err != nil {
		fmt.Println("failed to create topic:", err)
	}
	_ = c.Close()

	conn, err := connectkafka(topic, partition)
	if err != nil {
		fmt.Println("failed to connect to kafka:", err)
		return
	}
	writeMessages(conn, []string{"wave 1"})
	// readMessages(conn, 10, 10e3)
	// if err := conn.Close(); err != nil {
	// 	fmt.Println("failed to close connection:", err)
	// }
	log.Println("Kafka insert completed successfully")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    "wave",
		GroupID:  "wave", // Consumer group
		MinBytes: 10,
		MaxBytes: 10e3,
	})
	defer reader.Close()
	msg, err := reader.ReadMessage(context.Background())
	if err != nil {
		fmt.Println("failed to read message:", err)
	} else {
		fmt.Printf("message at offset %d: %s = %s\n", msg.Offset, string(msg.Key), string(msg.Value))
	}

	http.HandleFunc("/", list)
	http.HandleFunc("/createtable", createTableHandler)
	http.HandleFunc("/insert", insertDataHandler)
	http.HandleFunc("/createrestaurant", addNewRestaurantHandler)
	http.HandleFunc("/addmenuitem", addMenuItemHandler)
	http.HandleFunc("/getmenubyrestaurant", getMenuByRestaurantHandler)
	http.HandleFunc("/createdelivery", createDeliveryHandler)

	// order menu call api to java
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}

	// client := &http.Client{}
	// resp, err := client.Get("http://localhost:8081/")
	// if err != nil {
	// 	fmt.Println("Error making GET request:", err)
	// 	return
	// }
	// defer resp.Body.Close()
	// fmt.Println("Response Status:", resp.Status)
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println("Error reading response body:", err)
	// 	return
	// }
	// fmt.Println("Response Body:", string(body))
	// defer conn.Close(context.Background())
}
