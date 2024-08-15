package main

import (
	"encoding/json"
	"fmt"
	// "io"
	"log"
	"net/http"
	// "os"
	"time"

	// h "github.com/alurujawahar/tejimandi/httpRequest"
	order "github.com/alurujawahar/tejimandi/order"
	
	"github.com/gorilla/mux"
	
	bolt "go.etcd.io/bbolt"
)




const apiUrl = "localhost:8080"

var (
    db        *bolt.DB
    authToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Mjg3NTc4NzV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
    
)

func initDB() {
    var err error
    db, err = bolt.Open("stocks.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
    if err != nil {
        log.Fatalf("failed to open database: %v", err)
    }
}



func getStockHandler(w http.ResponseWriter, r *http.Request) {
    symbol := mux.Vars(r)["symbol"]

    var stock order.StockData
    err := db.View(func(tx *bolt.Tx) error {
        bucket := tx.Bucket([]byte("Stocks"))
        if bucket == nil {
            return fmt.Errorf("Bucket not found")
        }

        data := bucket.Get([]byte(symbol))
        if data == nil {
            return fmt.Errorf("Stock not found")
        }

        return json.Unmarshal(data, &stock)
    })

    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stock)
}

func setStockHandler(w http.ResponseWriter, r *http.Request) {
    var stock order.StockData
    if err := json.NewDecoder(r.Body).Decode(&stock); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    err := db.Update(func(tx *bolt.Tx) error {
        bucket, err := tx.CreateBucketIfNotExists([]byte("Stocks"))
        if err != nil {
            return fmt.Errorf("create bucket: %w", err)
        }

        data, err := json.Marshal(stock)
        if err != nil {
            return fmt.Errorf("failed to marshal stock data: %w", err)
        }
        if err := bucket.Put([]byte(stock.Token), data); err != nil {
            return fmt.Errorf("failed to put stock data into bucket: %w", err)
        }

        return nil
    })

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    fmt.Fprintf(w, "Stock data for %s set successfully!", stock.Token)
}

func setMultipleStocksHandler(w http.ResponseWriter, r *http.Request) {
    var stocks []order.StockData
    if err := json.NewDecoder(r.Body).Decode(&stocks); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    err := db.Update(func(tx *bolt.Tx) error {
        bucket, err := tx.CreateBucketIfNotExists([]byte("Stocks"))
        if err != nil {
            return fmt.Errorf("create bucket: %w", err)
        }

        for _, stock := range stocks {
            data, err := json.Marshal(stock)
            if err != nil {
                return fmt.Errorf("failed to marshal stock data: %w", err)
            }
            if err := bucket.Put([]byte(stock.Token), data); err != nil {
                return fmt.Errorf("failed to put stock data into bucket: %w", err)
            }
        }

        return nil
    })

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    fmt.Fprintf(w, "Stock data set successfully!")
}

func getAllStocksHandler(w http.ResponseWriter, r *http.Request) {
    var stocks []order.StockData

    err := db.View(func(tx *bolt.Tx) error {
        bucket := tx.Bucket([]byte("Stocks"))
        if bucket == nil {
            return fmt.Errorf("Bucket not found")
        }

        err := bucket.ForEach(func(k, v []byte) error {
            var stock order.StockData
            if err := json.Unmarshal(v, &stock); err != nil {
                return fmt.Errorf("failed to unmarshal stock data: %w", err)
            }
            stocks = append(stocks, stock)
            return nil
        })

        if err != nil {
            return fmt.Errorf("failed to iterate through bucket: %w", err)
        }

        return nil
    })

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stocks)
}


func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        fmt.Println("Token:", token)
        if token != authToken {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func main() {
	// stocksFilePath := "/Users/alurujawahar/Desktop/angel/tejimandi/stocks_partial.json"
	
	
	

	initDB()
    defer db.Close()

    r := mux.NewRouter()
    r.HandleFunc("/stock/{symbol}", getStockHandler).Methods("GET")
    r.HandleFunc("/stock", setStockHandler).Methods("POST")
    r.HandleFunc("/stocks", setMultipleStocksHandler).Methods("POST")
    r.HandleFunc("/stocks", getAllStocksHandler).Methods("GET")
	r.HandleFunc("/placeOrder", order.PlaceBulkOrder).Methods("POST")

	r.Use(authMiddleware)

	http.Handle("/", r)
    fmt.Println("Server listening on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
	
	//Get Authenticated
	// ABClient, _, _ := authenticate(filepath)
	hour, _, _ := time.Now().Clock()
	if (hour >= 9) && (hour <= 15) {
		if true {
			// order.PlaceBulkOrder(ABClient, stocksFilePath, "NSE")
		}
	} else {
		fmt.Println("Can't trade since out of market hours")
	}
}