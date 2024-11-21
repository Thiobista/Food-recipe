package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var usersDB = map[string]string{
	"testuser": "password123",
}

func GetUser(ctx context.Context, username string) (map[string]interface{}, error) {
	req := graphql.NewRequest(`
		query($username: String!) {
			users(where: {username: {_eq: $username}}) {
				id
				username
			}
		}
	`)
	req.Var("username", username)

	var resp map[string]interface{}
	err := h.client.Run(ctx, req, &resp)
	if err != nil {
		return nil, fmt.Errorf("error fetching user: %v", err)
	}

	return resp, nil
}

// Signup endpoint
func Signup(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	hasuraClient := NewHasuraClient("http://localhost:8080/v1/graphql")

	err = hasuraClient.InsertUser(context.Background(), user.Username, user.Password)
	if err != nil {
		http.Error(w, "Error inserting user to Hasura: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created successfully"))
}


// Login endpoint
func Login(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	storedPassword, exists := usersDB[user.Username]
	if !exists || storedPassword != user.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := "dummyJWT" // Replace with actual JWT generation logic
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// ProtectedRoute endpoint
func ProtectedRoute(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to the protected route!"))
}

// Default home handler
func Home(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server is running!"))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", Home).Methods("GET")
	r.HandleFunc("/signup", Signup).Methods("POST")
	r.HandleFunc("/login", Login).Methods("POST")
	r.HandleFunc("/protected-route", ProtectedRoute).Methods("GET")

	log.Fatal(http.ListenAndServe(":8081", r))
}
