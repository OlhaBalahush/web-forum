package middleware

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/OlhaBalahush/web-forum/functions/handlers"
)

type Limiter struct {
	ipCount map[string]int
	sync.Mutex
}

var limiter Limiter

func init() {
	limiter.ipCount = make(map[string]int)
}

// Limits the users(does it with ip address) requests to a maximum of ten, every request a request is added to the user
// every 2 seconds a request is removed from the user
// every 10 seconds all the requests are deleted
func Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//get the ip address of current user
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		limiter.Lock()
		count, ok := limiter.ipCount[ip]
		if !ok {
			limiter.ipCount[ip] = 0
		}
		if count > 30 {
			limiter.Unlock()
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		} else {
			limiter.ipCount[ip]++
		}
		time.AfterFunc(time.Second*10, func() {
			limiter.Lock()
			limiter.ipCount[ip]--
			limiter.Unlock()
		})
		if limiter.ipCount[ip] == 40 {
			//set it to 150, so the decrement timers will only decrease it to
			//100, and they stay blocked until the next timer resets it to 0
			limiter.ipCount[ip] = 50
			time.AfterFunc(time.Second*60, func() {
				limiter.Lock()
				limiter.ipCount[ip] = 0
				limiter.Unlock()
			})
		}
		limiter.Unlock()
		next.ServeHTTP(w, r)
	})
}

func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Refferer Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event
		// of a panic as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a
			// panic or not. If there has...
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500
				// Internal Server response.
				log.Println(err)
				handlers.RenderErrorPage(w, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
