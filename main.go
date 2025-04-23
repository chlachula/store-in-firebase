//https://gemini.google.com/app/fb0c1b03b98eab7c

package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/db"
	"github.com/gorilla/mux"

	//"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// Person struct to hold the form data
type Person struct {
	Name            string           `json:"name"`
	DateOfBirth     string           `json:"date_of_birth"`
	Deceased        bool             `json:"deceased"`
	DateOfDeath     string           `json:"date_of_death,omitempty"`
	Accomplishments []Accomplishment `json:"accomplishments"`
}

// Accomplishment struct
type Accomplishment struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
	Description string `json:"description"`
	Color       string `json:"color,omitempty"`
}

var firebaseClient *db.Client

//go:embed  form.html
var formBytes []byte

//go:embed  operations.html
var operationsBytes []byte

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<html>
			<head>
				<title>Home Page</title>
			</head>
			<body style="text:center;">
				<h1>Welcome to the Home Page!</h1>
				<a href="/operations">Operations</a><br>
				<a href="/person_add">Add Person</a><br>
				<a href="/person_list">List People</a><br>
				<a href="/abc">abc directory</a><br>
				/person/{key} ... to do ...<br>
			</body>
		</html>
	`))
}

func formPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received method:%s request to url:%s\n", r.Method, r.URL.Path) //
	w.Header().Set("Content-Type", "text/html")
	w.Write(formBytes)
}

func operationsPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received method:%s request to url:%s\n", r.Method, r.URL.Path) //
	w.Header().Set("Content-Type", "text/html")
	w.Write(operationsBytes)
}

func submitFormHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received method:%s request to url:%s\n", r.Method, r.URL.Path) //
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	/*
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			fmt.Printf("Error parsing form: %v\n", err) //
			log.Printf("Error parsing form: %v", err)
			return
		}
	*/
	var person Person
	err := json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		log.Printf("Error decoding request body: %v", err)
		return
	}
	/*
		var person Person
		person = updatedPerson

		person.Name = r.FormValue("updateName")
		person.DateOfBirth = r.FormValue("updateDob")
		person.Deceased = r.FormValue("deceased") == "on" // Checkbox value

		if person.Deceased {
			person.DateOfDeath = r.FormValue("dod")
		}

		// Handle accomplishments dynamically
		var accomplishments []Accomplishment
		for i := 0; ; i++ {
			startDate := r.FormValue(fmt.Sprintf("accomplishments[%d][start_date]", i))
			description := r.FormValue(fmt.Sprintf("accomplishments[%d][description]", i))

			if startDate == "" || description == "" {
				break // No more accomplishments
			}

			endDate := r.FormValue(fmt.Sprintf("accomplishments[%d][end_date]", i))
			color := r.FormValue(fmt.Sprintf("accomplishments[%d][color]", i))

			accomplishments = append(accomplishments, Accomplishment{
				StartDate:   startDate,
				EndDate:     endDate,
				Description: description,
				Color:       color,
			})
		}
		person.Accomplishments = accomplishments

		// Basic server-side validation (you can add more robust validation)
		if person.Name == "" || person.DateOfBirth == "" {
			http.Error(w, "Name and Date of Birth are required", http.StatusBadRequest)
			return
		}
		if person.Deceased && person.DateOfDeath == "" {
			http.Error(w, "Date of Death is required if deceased", http.StatusBadRequest)
			return
		}

		for _, acc := range person.Accomplishments {
			if acc.StartDate == "" || acc.Description == "" {
				http.Error(w, "Start Date and Description are required for each accomplishment", http.StatusBadRequest)
				return
			}
			// You could add date format validation here
		}
	*/
	// Store data in Firebase Realtime Database
	ref := firebaseClient.NewRef("people") // You can choose a different path
	_, err = ref.Push(r.Context(), person)
	if err != nil {
		http.Error(w, "Error saving data to Firebase", http.StatusInternalServerError)
		fmt.Printf("Error saving to Firebase: %#v\n", err) //
		log.Printf("Error saving to Firebase: %#v", err)
		return
	}

	fmt.Fprintf(w, "Data submitted successfully!\n %v \n", person)
	log.Printf("Data submitted successfully! %v", person)
}
func getPersonHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if key == "" {
		http.Error(w, "Missing person key", http.StatusBadRequest)
		return
	}

	ref := firebaseClient.NewRef(fmt.Sprintf("people/%s", key))
	var person Person
	err := ref.Get(r.Context(), &person)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading person with key %s: %v", key, err), http.StatusInternalServerError)
		log.Printf("Error reading person with key %s: %v", key, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(person)
	if err != nil {
		log.Printf("Error encoding person to JSON: %v", err)
	}
}
func updatePersonHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if key == "" {
		http.Error(w, "Missing person key", http.StatusBadRequest)
		return
	}

	var updatedPerson Person
	err := json.NewDecoder(r.Body).Decode(&updatedPerson)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		log.Printf("Error decoding request body: %v", err)
		return
	}

	ref := firebaseClient.NewRef(fmt.Sprintf("people/%s", key))
	err = ref.Set(r.Context(), updatedPerson) // Use Set to overwrite the existing data
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating person with key %s: %v", key, err), http.StatusInternalServerError)
		log.Printf("Error updating person with key %s: %v", key, err)
		return
	}

	fmt.Fprintf(w, "Person with key %s updated successfully!", key)
}
func deletePersonHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if key == "" {
		http.Error(w, "Missing person key", http.StatusBadRequest)
		return
	}

	ref := firebaseClient.NewRef(fmt.Sprintf("people/%s", key))
	err := ref.Delete(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting person with key %s: %v", key, err), http.StatusInternalServerError)
		log.Printf("Error deleting person with key %s: %v", key, err)
		return
	}

	fmt.Fprintf(w, "Person with key %s deleted successfully!", key)
}

func listPeopleHandler(w http.ResponseWriter, r *http.Request) {
	ref := firebaseClient.NewRef("people")
	var people map[string]Person // Use a map to hold key-value pairs

	err := ref.Get(r.Context(), &people)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading all people: %v", err), http.StatusInternalServerError)
		log.Printf("Error reading all people: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(people)
	if err != nil {
		log.Printf("Error encoding people to JSON: %v", err)
	}
}

func personsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("personsHandler-Received method:%s request to url:%s\n", r.Method, r.URL.Path) //
	if r.Method == http.MethodGet {
		vars := mux.Vars(r)
		key := vars["key"]
		if key == "home" {
			homeHandler(w, r)
		} else if key == "operations" {
			operationsPageHandler(w, r)
		} else if key == "list" {
			listPeopleHandler(w, r)
		} else if key == "add" {
			formPageHandler(w, r)
		} else {
			getPersonHandler(w, r)
		}
	} else if r.Method == http.MethodPost {
		submitFormHandler(w, r)
	} else if r.Method == http.MethodPut {
		updatePersonHandler(w, r)
	} else if r.Method == http.MethodDelete {
		deletePersonHandler(w, r)
	} else {
		http.Error(w, "Unexpected method", http.StatusMethodNotAllowed)
		return
	}
}

// login related
var firebaseAuthClient *auth.Client // Firebase Auth client

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This is a protected resource!")
}

// Example of middleware to verify Firebase ID token (basic)
func authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Assuming the token is in the format "Bearer <ID_TOKEN>"
		token := ""
		_, err := fmt.Sscan(authHeader, "Bearer", &token)
		if err != nil {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		// Verify the ID token
		ctx := context.Background()
		_, err = firebaseAuthClient.VerifyIDToken(ctx, token)
		if err != nil {
			http.Error(w, "Invalid ID token", http.StatusUnauthorized)
			log.Printf("Error verifying ID token: %v", err)
			return
		}

		// Token is valid, proceed to the next handler
		next.ServeHTTP(w, r)
	}
}
func main() {
	// Load environment variables from .env file
	/* let's not use it for now
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	*/

	// Initialize Firebase with a background context
	ctx := context.Background()
	config := &firebase.Config{
		DatabaseURL: os.Getenv("FIREBASE_DATABASE_URL"),
	}
	opt := option.WithCredentialsFile(os.Getenv("FIREBASE_CREDENTIALS_PATH"))
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	client, err := app.Database(ctx)
	if err != nil {
		log.Fatalf("error initializing database client: %v", err)
	}
	firebaseClient = client

	//login related
	// Initialize Firebase Auth client
	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error initializing auth client: %v", err)
	}
	firebaseAuthClient = authClient

	// Set up HTTP routes
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/operations", operationsPageHandler)
	r.HandleFunc("/person_add", formPageHandler)
	r.HandleFunc("/person_list", listPeopleHandler)
	r.HandleFunc("/persons/{key}", personsHandler).Methods("GET", "PUT", "DELETE", "POST")
	r.HandleFunc("/person/{key}", getPersonHandler).Methods("GET")
	r.HandleFunc("/person/{key}", updatePersonHandler).Methods("PUT")    // New route for updating
	r.HandleFunc("/person/{key}", deletePersonHandler).Methods("DELETE") // New route for deleting
	r.HandleFunc("/person_submit", submitFormHandler).Methods("POST")
	//r.PathPrefix("/").Handler(http.FileServer(http.Dir("./"))) // Serve static files (index.html)
	r.HandleFunc("/protected", authenticate(protectedHandler)).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	fmt.Printf("Server listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

/*
Important Considerations for Backend Authentication:

* Firebase Admin SDK Credentials: Ensure your service account key file has the necessary permissions for Firebase Authentication.
* Token Transmission: You'll need a way for your frontend to send the Firebase ID token to your backend
  (e.g., in the Authorization header of your fetch requests after a successful login).
* Error Handling: Implement robust error handling for token verification failures.
* Security: Always handle ID tokens securely and follow Firebase security best practices.
This comprehensive example provides the basic steps to add user login to your Firebase-backed web application.
Remember to replace the placeholder Firebase configuration with your actual credentials.
For more advanced authentication scenarios and backend protection,
refer to the Firebase documentation for the Web SDK and the Go Admin SDK.
*/
