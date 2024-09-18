package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var redisClient *redis.Client

func main() {
	redisURL := os.Getenv("REDIS_URL")

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal(err)
	}
	redisClient = redis.NewClient(opt)

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "Hello, HTTP!\n")
	}).Methods("GET")

	r.HandleFunc("/emit", emitHandler).Methods("POST")

	port := os.Getenv("PORT")
	log.Println("API server starting on ", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}

type Message struct {
	Time  time.Time `json:"time"`
	Event string    `json:"event"`
	Room  string    `json:"room"`
	Data  any       `json:"data"`
}

func (m *Message) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}
func (m *Message) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func emitHandler(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	room := req["room"].(string)
	content := req["content"]

	payload := map[string]any{
		"event": "new_message",
		"room":  room,
		"data": map[string]any{
			"time":    time.Now().UTC(),
			"content": content,
		},
	}

	fmt.Println("go-api payload", payload)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = redisClient.Publish(r.Context(), "socket.io", jsonPayload).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
