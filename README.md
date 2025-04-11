# Drive Management System

Welcome to the **Drive Management System**! A lightweight file management system to create, upload, rename, delete files (of given types) or folders.

## Features
- **User Authentication:** Secure registration and login with JWT.
- **User Activity Log:** User Activity is recorded like creating, updating, deleting Drive or registering, logging, account activation .
- **Drive Management:** Create, read, update, delete (soft delete) file and folders.
- **Logging:** Using `Lagrus` for structured logging.
- **Rate Limiting:** Goroutine-based rate limiter.
- **Server Error Handling:** Env-based maintenance mode.
- **Context Middleware:** Each request has a **5-second timeout** for better resource management.
- **Database Migrations:** Managed migration using GORM.
- **Directory Listing:** View the contents of your directories.

## Technologies Used
- **GoLang:** Backend development
- **Gin:** HTTP web framework
- **MySQL:** Database management
- **Bcrypt:** Secure password hashing
- **JWT:** Token-based authentication
- **GORM:** ORM for database interactions
- **Lagrus:** Logging system
- **SecureHeader:** Additional security headers
- **HTML/CSS/JS:** – UI with client-side form handling and validation

## API Endpoints

### **User Authentication**
- `POST /register` - Register a new user
- `GET /activation_token/:token` - Activate user account
- `POST /login` - Authenticate and receive JWT token

### **Drive Management**
- `GET /drive` - List all the files and folders after authentication
- `GET /drive/img.png` - Get a single img if exist img.png
- `POST /drive/create` - Create a new folder
- `POST /drive/upload` - Create a new file
- `PUT /drive/rename` - Rename a file or folder
- `DELETE /tasks/delete/:id` - Soft delete a task

## Getting Started

### **Prerequisites**
- Install **GoLang** ([Download](https://golang.org/dl/))
- Install **MySQL** ([Download](https://www.mysql.com/download/))

### **Installation**
1. Clone the repository:
   ```sh
   git clone github.com/iamgak/go-drive
   ```
2. Navigate to the project directory:
   ```sh
   cd go-drive
   ```
3. Install dependencies:
   ```sh
   go mod tidy
   ```
4. Create MySQL database:
   ```sh
   mysql -u root -p -e "CREATE DATABASE go_task;"
   ```
5. Run the server:
   ```sh
   go run .
   ```

## Context Middleware (5-Second Timeout)
To prevent long-running requests and manage resources efficiently, a **global middleware** enforces a **5-second timeout** for each API request:
```go
func TimeoutMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
        defer cancel()
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}
```
This middleware is applied globally in `main.go`:
```go
r := gin.Default()
r.Use(TimeoutMiddleware())
```

## Usage Examples
- **Register a new user:** `POST https://localhost:8000/register`
- **Login:** `POST https://localhost:8000/login`
- **Fetch all file and folder:** `GET https://localhost:8000/drive`
- **Get a folder or file by Name:** `GET https://localhost:8000/drive/:name`
- **Create a folder:** `POST https://localhost:8000/drive/create`
- **Create a file in a folder:** `POST https://localhost:8000/drive/upload`
- **rename a file or folder:** `PUT https://localhost:8000/drive/rename`

Example requests:
```sh
curl -X GET "localhost:8080/login"
curl -X GET "localhost:8080/activation_token/{verification_token}"
```

## Contributing
Contributions are welcome! Fork the repository and submit a pull request with your changes.

