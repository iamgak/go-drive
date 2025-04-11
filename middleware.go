package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/iamgak/go-drive/models"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

func secureHeaders() gin.HandlerFunc {
	return (func(c *gin.Context) {
		// css was not loading SOP
		// c.Header("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		// c.Header("Referrer-Policy", "origin-when-cross-origin")
		// c.Header("X-Content-Type-Options", "nosniff")
		// c.Header("X-Frame-Options", "deny")
		// c.Header("X-XSS-Protection", "0")
		c.Next()
	})
}
func (app *Application) LoginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("ldata")
		if err != nil || cookie.Value == "" {
			app.sendJSONResponse(c.Writer, http.StatusUnauthorized, "Access Denied")
			app.Logger.Warning("Missing or empty ldata cookie")
			c.Abort()
			return
		}

		tokenString := cookie.Value

		err = godotenv.Load()
		if err != nil {
			app.sendJSONResponse(c.Writer, http.StatusInternalServerError, "Internal Server Error")
			app.Logger.Errorf("Error loading env: %v", err)
			c.Abort()
			return
		}

		SIGNING_KEY := os.Getenv("SIGNING_KEY")
		if SIGNING_KEY == "" {
			app.sendJSONResponse(c.Writer, http.StatusInternalServerError, "Signing key not found")
			app.Logger.Error("Missing SIGNING_KEY in env")
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &models.MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(SIGNING_KEY), nil
		})

		if err != nil {
			app.sendJSONResponse(c.Writer, http.StatusUnauthorized, "Invalid Token")
			app.Logger.Error("Token parse error:", err)
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*models.MyCustomClaims); ok && token.Valid {
			app.UserID = claims.UserID
			app.Email = claims.Email
			app.BaseDir = fmt.Sprintf("/home/iamgak/Desktop/info/assignment/go-task-github/drive/%d/", app.UserID)
			app.isAuthenticated = true
			c.Next()
		} else {
			app.Logger.Warning("Invalid token or claims")
			app.sendJSONResponse(c.Writer, http.StatusUnauthorized, "Invalid Token")
			c.Abort()
			return
		}
	}
}

func (app *Application) rateLimiter() gin.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Declare a mutex and a map to hold the clients' IP addresses and rate limiters.
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			// it will run until code run but take break every minute laziness
			time.Sleep(time.Minute)
			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place.
			mu.Lock()
			// Loop through all clients. If they haven't been seen within the last three
			// minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip, _, err := net.SplitHostPort(c.ClientIP())
		if err != nil {
			ip := net.ParseIP(c.ClientIP())
			if ip == nil {
				app.Logger.Warn("Invalid IP address :", c.ClientIP())
				app.ServerError(c.Writer, err)
				return
			}
		}
		// Lock the mutex to prevent this code from being executed concurrently.

		mu.Lock()
		if _, found := clients[ip]; !found {
			// Create and add a new client struct to the map if it doesn't already exist.
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Limit(5), 3),
			}
		}

		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.CustomError(c.Writer, http.StatusTooManyRequests, "Too, many request. Rate Limit Exceed")
			return
		}

		mu.Unlock()
	}
}

func (app *Application) TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request timeout"})
			c.Abort()
			return
		}
	}
}

// just in case in future maintainance needed
func MaintenanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetHeader("X-User-Role") // only open for me

		if os.Getenv("SERVER_STATUS") == "maintenance" && userRole != "admin" {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "The server is currently under maintenance. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
