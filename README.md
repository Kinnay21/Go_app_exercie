**Bike Battery Charging API**
===========================

This is a Go RESTful API for managing bike battery charging at charging stations. The API uses the Gin framework and MySQL database.

## **Getting Started**

Follow the instructions below to get a copy of the project up and running on your local machine for development and testing purposes.
Prerequisites

To run this project, you need to have the following software installed on your machine:
* Go 1.16 or later
* MySQL

## **Usage**

The following endpoints are available:

### **`GET /batteries`**
Returns a list of all batteries.

### **`GET /batteries/:id`**
Returns the battery with the given ID.

### **`POST /batteries`**
Adds a new battery to the database. Expects a JSON payload with the following fields:

* ``id (string)``: The battery's unique ID.
* ``level (float)``: The battery's current charge level.
* ``is_charging (boolean)``: Whether the battery is currently charging.
* ``charging_speed (float)``: The speed at which the battery is charging.

### **`PUT /batteries/:id`**
Updates the battery with the given ID. Expects a JSON payload with the same fields as the POST /batteries endpoint.

### **`POST /charge/:id`**
Puts the battery with the given ID into charging mode. Returns a message indicating whether charging was started or if it was already in progress.

### **`DELETE /charge/:id`**
Stops charging the battery with the given ID.

## **Built With**

    Go - Programming language
    Gin - HTTP web framework
    MySQL - Relational database management system


## **Setup the Database**

Install MySQL on your machine or server, if it is not already installed. You can download the MySQL Community Server from the official website: https://dev.mysql.com/downloads/mysql/.

Create a new database for the application to use. You can do this using the MySQL command line client or a graphical tool such as phpMyAdmin. For example, to create a database named `charging_stations`, you can run the following command:

```sql:
CREATE DATABASE charging_stations;
```

Create a new table to store data about bike batteries. You can use the following SQL statement to create a `batteries` table with four columns: `id`, `level`, `is_charging`, and `charging_speed`.

```sql:
CREATE TABLE batteries (
    id VARCHAR(50) PRIMARY KEY,
    level FLOAT(4,2) NOT NULL,
    is_charging BOOL NOT NULL,
    charging_speed FLOAT(4,2) NOT NULL
);
``` 

Insert some sample data into the `batteries` table using the following SQL statements:


```sql:
INSERT INTO batteries (id, level, is_charging, charging_speed) VALUES
('001', 80.0, false, 5.0),
('002', 50.0, false, 2.0),
('003', 30.0, false, 1.0),
('004', 90.0, false, 0.5);
```

Open the Go code in your preferred code editor, and locate the following line:


```go:
db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/charging_stations")
```

Replace `root` and `password` with your MySQL username and password, and `charging_stations` with the name of the database you created in step 2.

Save the changes to the file, and run the Go program using the `go run` command. If everything is set up correctly, you should see output similar to the following:

```csharp:
    [GIN-debug] Listening and serving HTTP on :8080
```

This means that the server is running and ready to accept requests. You can now use a tool like `curl` or `Postman` to test the API endpoints.