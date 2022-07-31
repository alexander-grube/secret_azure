package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"secret-azure/model"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func postSecretHandler(rdb *redis.Client) http.HandlerFunc {
    fn := func(w http.ResponseWriter, r *http.Request) {
    secret := model.Secret{}
    err := json.NewDecoder(r.Body).Decode(&secret)
    log.Println(secret)
    if err != nil {
        log.Println(err)
        http.Error(w, "Error decoding secret", http.StatusBadRequest)
        return
    }
    id := uuid.New().String()
    secret.ID = id

    s, err := json.Marshal(secret)
    if err != nil {
        log.Println(err)
        http.Error(w, "Error marshalling secret", http.StatusInternalServerError)
        return
    }

    err = rdb.Set(r.Context(), secret.ID, s, secret.TTL).Err()
    if err != nil {
        log.Println(err)
        http.Error(w, "Error saving secret", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusCreated)
    fmt.Fprintf(w, secret.ID)
}
    return fn

}


func main() {
    rdbOptions, err := redis.ParseURL(os.Getenv("REDIS_DB"))
    if err != nil {
        log.Fatal(err)
    }
    rdb := redis.NewClient(rdbOptions)
    defer rdb.Close()
    listenAddr := ":8080"
    if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
        listenAddr = ":" + val
    }

    r := mux.NewRouter()
    r.HandleFunc("/api/secret", postSecretHandler(rdb)).Methods("POST")
    log.Printf("About to listen on %s. Go to https://127.0.0.1%s/", listenAddr, listenAddr)
    log.Fatal(http.ListenAndServe(listenAddr, r))
}