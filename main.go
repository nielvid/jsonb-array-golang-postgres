package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lib/pq"
)

type Order struct {
	ID    int    `json:"id"`
	Items []Item `json:"items"`
}

type Item struct {
	ID    uint8  `json:"id"`
	Name  string `json:"name"`
	Price uint8  `json:"price"`
}

func (o Order) Value() (driver.Value, error) {
	return json.Marshal(o)
}
func (o *Order) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}
	return json.Unmarshal(b, o)
}

// Function to insert a record into the orders table
func insertOrder(db *sql.DB, order Order) error {
	// Convert the slice of Item structs to JSONB[]
	jsonbItems := make([]string, len(order.Items))
	for i, item := range order.Items {
		jsonItem, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item %v: %v", item, err)
		}
		jsonbItems[i] = string(jsonItem)
	}

	// Insert the order into the database
	query := `INSERT INTO orders (id, items) VALUES ($1, $2) RETURNING id`
	err := db.QueryRow(query, order.ID, pq.Array(jsonbItems)).Scan(&order.ID)
	if err != nil {
		return fmt.Errorf("failed to insert order: %v", err)
	}

	fmt.Printf("Inserted order with ID: %d\n", order.ID)
	return nil
}


// Function to retrieve a record from the orders table by ID
func getOrder(db *sql.DB, id int) (Order, error) {
    var order Order

    // Query to retrieve the order by ID
    query := `SELECT id, items FROM orders WHERE id = $1`
    row := db.QueryRow(query, id)

    var itemsData []string
    if err := row.Scan(&order.ID, pq.Array(&itemsData)); err != nil {
        return order, fmt.Errorf("failed to retrieve order: %v", err)
    }

    // Unmarshal each JSONB[] item into the Items slice
    for _, itemData := range itemsData {
        var item Item
        if err := json.Unmarshal([]byte(itemData), &item); err != nil {
            return order, fmt.Errorf("failed to unmarshal item: %v", err)
        }
        order.Items = append(order.Items, item)
    }

    return order, nil
}


func getAllOrders(db *sql.DB) ([]Order, error) {
    var orders []Order

    // Query to retrieve the order by ID
    query := `SELECT id, items FROM orders`
    rows, _ := db.Query(query)

    for rows.Next() {
		  var order Order
        var itemsData []string
        if err := rows.Scan(&order.ID, pq.Array(&itemsData)); err != nil {
            return nil, fmt.Errorf("failed to retrieve order: %v", err)
        }
		// Unmarshal each JSONB[] item into the Items slice
    for _, itemData := range itemsData {
        var item Item
        if err := json.Unmarshal([]byte(itemData), &item); err != nil {
            return nil, fmt.Errorf("failed to unmarshal item: %v", err)
        }
		order.Items = append(order.Items, item)
		   
    }
	// Append the order to the list of orders
        orders = append(orders, order)
}
return orders, nil

}
func main() {
	// Connect to the PostgreSQL database
	connStr := "user=user dbname=db password=password host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	
	order4 := Order{
		ID: 4,
		Items: []Item{
			{ID: 1, Name: "milk", Price: 50},
			{ID: 2, Name: "sugar", Price: 20},
		},
	}

	// Insert the order
	if err := insertOrder(db, order4); err != nil {
		log.Fatalf("failed to insert order: %v", err)
	}


	 // Example: Retrieve an order with ID 1
    order, err := getOrder(db, 1)
    if err != nil {
        log.Fatalf("failed to retrieve order: %v", err)
    }

    fmt.Printf("Retrieved Order: %+v\n", order)

	orders, err := getAllOrders(db)

	 if err != nil {
        log.Fatalf("failed to retrieve order: %v", err)
    }

     fmt.Printf("Retrieved Order: %+v\n", orders)

	

	
}
