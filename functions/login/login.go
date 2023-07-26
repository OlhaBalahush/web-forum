package login

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// Your serverless function logic here
	fmt.Fprintln(w, "Hello from the serverless function!")
}
