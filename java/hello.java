import com.sun.net.httpserver.HttpServer;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpExchange;
import java.net.InetSocketAddress;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.sql.*;
import org.json.JSONObject;

public class hello {
    public static void main(String[] args) throws IOException{
        System.out.println("Hello, World!");
        HttpServer server = HttpServer.create(new InetSocketAddress("0.0.0.0", 8085), 0);
        server.createContext("/hello", new HelloHandler());
        server.createContext("/delivery", new DeliveryHandler());

        server.setExecutor(null); // creates a default executor
        server.start();
        System.out.println("Server started on port 8085");


    }
     static class DeliveryHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange exchange) throws IOException {
            if (!"POST".equals(exchange.getRequestMethod())) {
                exchange.sendResponseHeaders(405, -1);
                return;
            }

            // Read from environment variables with defaults
            String dbHost = System.getenv("DB_HOST");
            if (dbHost == null || dbHost.isEmpty()) {
                dbHost = "localhost";
            }

            String dbPort = System.getenv("DB_PORT");
            if (dbPort == null || dbPort.isEmpty()) {
                dbPort = "5432";
            }

            String username = System.getenv("DB_USER");
            if (username == null || username.isEmpty()) {
                username = "postgres";
            }

            String password = System.getenv("DB_PASSWORD");
            if (password == null || password.isEmpty()) {
                password = "mysecretpassword";
            }

            String dbName = System.getenv("DB_NAME");
            if (dbName == null || dbName.isEmpty()) {
                dbName = "postgres";
            }

            String jdbcURL = "jdbc:postgresql://" + dbHost + ":" + dbPort + "/" + dbName;

            try {
                // Read request body
                InputStreamReader isr = new InputStreamReader(exchange.getRequestBody(), StandardCharsets.UTF_8);
                BufferedReader br = new BufferedReader(isr);
                StringBuilder sb = new StringBuilder();
                String line;
                while ((line = br.readLine()) != null) {
                    sb.append(line);
                }

                JSONObject json = new JSONObject(sb.toString());
                
                String restaurantName = json.getString("restaurant_name");
                int restaurantId = json.getInt("restaurant_id");
                int orderId = json.getInt("order_id");
                String deliveryName = json.getString("delivery_name");
                String deliveryPhone = json.optString("delivery_phone", "");
                String deliveryLocation = json.getString("delivery_location");
                String customerName = json.getString("customer_name");
                String customerPhone = json.optString("customer_phone", "");
                double totalAmount = json.getDouble("total_amount");

                Class.forName("org.postgresql.Driver");
                Connection connection = DriverManager.getConnection(jdbcURL, username, password);
                
                // Insert into orders table
                String orderSQL = "INSERT INTO orders (id, restaurant_id, restaurant_name, customer_name, customer_phone, total_amount) VALUES (?, ?, ?, ?, ?, ?)";
                PreparedStatement orderStmt = connection.prepareStatement(orderSQL);
                orderStmt.setInt(1, orderId);
                orderStmt.setInt(2, restaurantId);
                orderStmt.setString(3, restaurantName);
                orderStmt.setString(4, customerName);
                orderStmt.setString(5, customerPhone);
                orderStmt.setDouble(6, totalAmount);
                orderStmt.executeUpdate();
                System.out.println("Order inserted successfully!");

                // Insert into deliveries table
                String deliverySQL = "INSERT INTO deliveries (order_id, delivery_name, delivery_phone, delivery_location) VALUES (?, ?, ?, ?)";
                PreparedStatement deliveryStmt = connection.prepareStatement(deliverySQL);
                deliveryStmt.setInt(1, orderId);
                deliveryStmt.setString(2, deliveryName);
                deliveryStmt.setString(3, deliveryPhone);
                deliveryStmt.setString(4, deliveryLocation);
                deliveryStmt.executeUpdate();
                System.out.println("Delivery inserted successfully!");

                // Notify delivery API
                try {
                    String pythonServiceURL = System.getenv("PYTHON_SERVICE_URL");
                    if (pythonServiceURL == null || pythonServiceURL.isEmpty()) {
                        pythonServiceURL = "http://localhost:8010";
                    }
                    
                    java.net.URL url = URI.create(pythonServiceURL + "/deliveryinitiated").toURL();
                    java.net.HttpURLConnection conn = (java.net.HttpURLConnection) url.openConnection();
                    conn.setRequestMethod("POST");
                    conn.setRequestProperty("Content-Type", "application/json");
                    conn.setDoOutput(true);
                    
                    JSONObject notificationPayload = new JSONObject();
                    notificationPayload.put("orderId", "ORD" + orderId);
                    notificationPayload.put("customerName", customerName);
                    notificationPayload.put("deliveryAddress", deliveryLocation);
                    notificationPayload.put("phoneNumber", customerPhone);
                    notificationPayload.put("items", new org.json.JSONArray());
                    notificationPayload.put("totalAmount", totalAmount);
                    
                    OutputStream notifyOs = conn.getOutputStream();
                    notifyOs.write(notificationPayload.toString().getBytes(StandardCharsets.UTF_8));
                    notifyOs.flush();
                    notifyOs.close();
                    
                    int notifyResponseCode = conn.getResponseCode();
                    System.out.println("Notification API response code: " + notifyResponseCode);
                    conn.disconnect();
                } catch (Exception notifyEx) {
                    System.err.println("Failed to notify delivery API: " + notifyEx.getMessage());
                }
                String response = "{\"status\": \"success\", \"message\": \"Delivery created successfully\"}";
                exchange.getResponseHeaders().set("Content-Type", "application/json");
                exchange.sendResponseHeaders(200, response.length());
                OutputStream os = exchange.getResponseBody();
                os.write(response.getBytes());
                os.close();

            } catch (Exception e) {
                e.printStackTrace();
                String errorResponse = "{\"status\": \"error\", \"message\": \"" + e.getMessage() + "\"}";
                exchange.sendResponseHeaders(500, errorResponse.length());
                OutputStream os = exchange.getResponseBody();
                os.write(errorResponse.getBytes());
                os.close();
            }
        }
    }
    static class HelloHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange exchange) throws IOException {
            // Read from environment variables with defaults
            String dbHost = System.getenv("DB_HOST");
            if (dbHost == null || dbHost.isEmpty()) {
                dbHost = "localhost";
            }

            String dbPort = System.getenv("DB_PORT");
            if (dbPort == null || dbPort.isEmpty()) {
                dbPort = "5432";
            }

            String username = System.getenv("DB_USER");
            if (username == null || username.isEmpty()) {
                username = "postgres";
            }

            String password = System.getenv("DB_PASSWORD");
            if (password == null || password.isEmpty()) {
                password = "mysecretpassword";
            }

            String dbName = System.getenv("DB_NAME");
            if (dbName == null || dbName.isEmpty()) {
                dbName = "postgres";
            }

            String jdbcURL = "jdbc:postgresql://" + dbHost + ":" + dbPort + "/" + dbName;

        try {
            // Load the PostgreSQL JDBC driver
            Class.forName("org.postgresql.Driver");

            // Establish the connection
            Connection connection
                = DriverManager.getConnection(
                    jdbcURL, username, password);
            System.out.println(
                "Connected to PostgreSQL database!");

            // Create a statement
            Statement statement
                = connection.createStatement();

            // Create a table if not exists
            String createTableSQL
                = "CREATE TABLE IF NOT EXISTS userstest (id SERIAL PRIMARY KEY, name VARCHAR(50), email VARCHAR(50))";
            statement.execute(createTableSQL);
            System.out.println("Table 'userstest' created!");

            // Insert a row into the table
            String insertSQL
                = "INSERT INTO userstest (name, email) VALUES ('John Doe', 'john.doe@example.com')";
            statement.executeUpdate(insertSQL);
            System.out.println(
                "Inserted data into 'userstest' table!");

            // Retrieve data from the table
            String selectSQL = "SELECT * FROM userstest";
            ResultSet resultSet
                = statement.executeQuery(selectSQL);

            while (resultSet.next()) {
                System.out.println(
                    "User ID: " + resultSet.getInt("id")
                    + ", Name: "
                    + resultSet.getString("name")
                    + ", Email: "
                    + resultSet.getString("email"));
            }

            // Close the connection
            connection.close();
            System.out.println("Connection closed.");
        }
        catch (Exception e) {
            e.printStackTrace();
        }
            String response = "Hello, World!";
            exchange.sendResponseHeaders(200, response.length());
            OutputStream os = exchange.getResponseBody();
            os.write(response.getBytes());  
            os.close();
        }
    }

    
}
