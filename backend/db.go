package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var jwtKey = []byte("super_secret_cyber_key")

// Models
type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Password string `gorm:"not null" json:"password"`
	Role     string `gorm:"default:'guest'" json:"role"` // admin, editor, guest
}

type Article struct {
	ID      string `gorm:"primaryKey" json:"id"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Content string `json:"content"`
}

type Project struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	Tech        string `json:"tech"`
	LinkDemo    string `json:"link_demo"`
	LinkGit     string `json:"link_git"`
	Category    string `json:"category"`
	Thumbnail   string `json:"thumbnail"`
}

type Contact struct {
	gorm.Model
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

type CTFScore struct {
	gorm.Model
	Username string `json:"username"`
	Score    int    `json:"score"`
}

type Analytics struct {
	gorm.Model
	Event string `gorm:"uniqueIndex" json:"event"`
	Count int    `json:"count"`
}

type Skill struct {
	gorm.Model
	Name  string `json:"name"`
	Level int    `json:"level"`
}

type LoginLog struct {
	gorm.Model
	Username  string `json:"username"`
	IPAddress string `json:"ip_address"`
	Status    string `json:"status"` // success, failed
}

type SystemLog struct {
	gorm.Model
	Message string `json:"message"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func loadEnv() {
	paths := []string{".env", "../.env"}
	for _, p := range paths {
		content, err := os.ReadFile(p)
		if err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					// Strip quotes
					if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
						val = val[1 : len(val)-1]
					} else if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") {
						val = val[1 : len(val)-1]
					}
					os.Setenv(key, val)
				}
			}
			break
		}
	}
}

func InitDB() error {
	loadEnv()

	var err error
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = "postgresql://postgres.cwxhumbjabgtpfklopjk:Local50n%24%5B%5D%7Cid3%3Dc%23%21Ter%24%2B%2B%7B%7D@aws-1-ap-northeast-2.pooler.supabase.com:6543/postgres?sslmode=require&prepare_threshold=0"
	}
	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Clean and drop old tables to ensure GORM schema matches models perfectly
	if err := db.AutoMigrate(&User{}, &Article{}, &Project{}, &Contact{}, &CTFScore{}, &Analytics{}, &LoginLog{}, &Skill{}, &SystemLog{}); err != nil {
		return fmt.Errorf("migration step 1 failed: %w", err)
	}

	// Migrate the schema
	if err := db.AutoMigrate(&User{}, &Article{}, &Project{}, &Contact{}, &CTFScore{}, &Analytics{}, &LoginLog{}, &Skill{}, &SystemLog{}); err != nil {
		return fmt.Errorf("migration step 2 failed: %w", err)
	}

	// Seed Admin
	var admin User
	if err := db.Where("LOWER(username) = LOWER(?)", "Novant").First(&admin).Error; err != nil {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("Novant123"), bcrypt.DefaultCost)

		if err := db.Create(&User{
			Username: "Novant",
			Email:    "novant@portfolio.local",
			Password: string(hashed),
			Role:     "admin",
		}).Error; err != nil {
			return fmt.Errorf("seeding admin failed: %w", err)
		}
	} else {
		if admin.Role != "admin" {
			admin.Role = "admin"
			db.Save(&admin)
		}
	}

	// Seed Projects if empty
	var projectCount int64
	db.Model(&Project{}).Count(&projectCount)
	if projectCount == 0 {
		if err := db.Create(&Project{Title: "Website Portfolio", Description: "Website portofolio interaktif dengan tema Hacker Terminal, dilengkapi dengan easter eggs, mini-games, dan CMS sederhana.", Tech: "HTML, CSS, Vanilla JS, Web Audio API", Category: "web"}).Error; err != nil {
			return fmt.Errorf("seeding projects failed: %w", err)
		}
		db.Create(&Project{Title: "Mobile App", Description: "Aplikasi mobile untuk pemantauan data sistem secara real-time dengan antarmuka futuristik.", Tech: "React Native, Node.js, WebSocket", LinkDemo: "https://responsi-farid.vercel.app/", Category: "app"})
	} else {
		// Update existing Mobile App project in case it's already seeded
		db.Model(&Project{}).Where("title = ?", "Mobile App").Update("link_demo", "https://responsi-farid.vercel.app/")
	}

	// Seed Skills if empty
	var skillCount int64
	db.Model(&Skill{}).Count(&skillCount)
	if skillCount == 0 {
		if err := db.Create(&Skill{Name: "⚡ JavaScript", Level: 85}).Error; err != nil {
			return fmt.Errorf("seeding skills failed: %w", err)
		}
		db.Create(&Skill{Name: "🌐 HTML & CSS", Level: 90})
		db.Create(&Skill{Name: "🧠 Problem Solving", Level: 80})
		db.Create(&Skill{Name: "🤝 Team Work", Level: 85})
	}

	// Migrate from data.json if necessary
	migrateOldData()
	return nil
}

func migrateOldData() {
	var count int64
	db.Model(&User{}).Count(&count)
	// If only admin exists, maybe we can migrate
	if count <= 1 {
		importDataJSON()
	}
}

func importDataJSON() {
	importFile := "data.json"
	data, err := os.ReadFile(importFile)
	if err != nil {
		return // Ignore if file doesn't exist
	}

	var oldData struct {
		Users    []map[string]interface{} `json:"users"`
		Articles []Article                `json:"articles"`
	}

	if err := json.Unmarshal(data, &oldData); err == nil {
		for _, u := range oldData.Users {
			username, _ := u["username"].(string)
			password, _ := u["password"].(string)
			if username != "" && password != "" {
				var existing User
				if db.Where("username = ?", username).First(&existing).Error != nil {
					hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
					email := fmt.Sprintf("%s@portfolio.local", username)
					db.Create(&User{Username: username, Email: email, Password: string(hashedPassword), Role: "guest"})
				}
			}
		}
		for _, a := range oldData.Articles {
			var existing Article
			if db.Where("id = ?", a.ID).First(&existing).Error != nil {
				db.Create(&a)
			}
		}
		fmt.Println("[SYSTEM] Migrated data.json to PostgreSQL database.")
	}
}

// Middleware Helper
func verifyJWT(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing token")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid token format")
	}

	tokenStr := parts[1]
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func HandleArticles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		var articles []Article
		db.Find(&articles)
		json.NewEncoder(w).Encode(articles)
	} else {
		// Protected routes
		claims, err := verifyJWT(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodPost {
			if claims.Role != "admin" && claims.Role != "editor" {
				http.Error(w, "Forbidden: Admins or Editors only", http.StatusForbidden)
				return
			}
			var newArticle Article
			if err := json.NewDecoder(r.Body).Decode(&newArticle); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			db.Save(&newArticle)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})

		} else if r.Method == http.MethodDelete {
			if claims.Role != "admin" {
				http.Error(w, "Forbidden: Admins only", http.StatusForbidden)
				return
			}
			id := r.URL.Query().Get("id")
			db.Delete(&Article{}, "id = ?", id)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func HandleUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		claims, err := verifyJWT(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if claims.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		var users []User
		db.Find(&users)
		for i := range users {
			users[i].Password = "[ENCRYPTED]"
		}
		json.NewEncoder(w).Encode(users)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user User
	if err := db.Where("LOWER(username) = LOWER(?)", creds.Username).First(&user).Error; err != nil {
		db.Create(&LoginLog{Username: creds.Username, IPAddress: r.RemoteAddr, Status: "failed"})
		recordAnalytics("login_failed")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Compare bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		db.Create(&LoginLog{Username: user.Username, IPAddress: r.RemoteAddr, Status: "failed"})
		recordAnalytics("login_failed")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Log successful login
	db.Create(&LoginLog{Username: user.Username, IPAddress: r.RemoteAddr, Status: "success"})
	recordAnalytics("user_login")
	broadcastLog("[AUTH] User logged in: " + user.Username)

	json.NewEncoder(w).Encode(map[string]string{
		"status":   "success",
		"token":    tokenString,
		"username": user.Username,
		"role":     user.Role,
	})
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newUser User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate email and username uniqueness
	var existing User
	if err := db.Where("LOWER(username) = LOWER(?)", newUser.Username).First(&existing).Error; err == nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}
	if err := db.Where("email = ?", newUser.Email).First(&existing).Error; err == nil {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	// Default role
	newUser.Role = "guest"

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error encryption password", http.StatusInternalServerError)
		return
	}
	newUser.Password = string(hashed)

	if err := db.Create(&newUser).Error; err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	recordAnalytics("user_register")
	broadcastLog("[AUTH] New user registered: " + newUser.Username)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func HandleAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		var payload struct {
			Event string `json:"event"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
			recordAnalytics(payload.Event)
			json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
		}
	} else if r.Method == http.MethodGet {
		claims, err := verifyJWT(r)
		if err != nil || claims.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		var analytics []Analytics
		db.Find(&analytics)
		json.NewEncoder(w).Encode(analytics)
	}
}

func HandleLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var scores []CTFScore
	db.Order("score desc, created_at asc").Limit(10).Find(&scores)
	json.NewEncoder(w).Encode(scores)
}

func recordAnalytics(event string) {
	var analytic Analytics
	if db.Where("event = ?", event).First(&analytic).Error != nil {
		db.Create(&Analytics{Event: event, Count: 1})
	} else {
		analytic.Count++
		db.Save(&analytic)
	}
}

func HandleContact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newContact Contact
	if err := json.NewDecoder(r.Body).Decode(&newContact); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.Create(&newContact).Error; err != nil {
		http.Error(w, "Error saving contact", http.StatusInternalServerError)
		return
	}

	recordAnalytics("contact_submit")
	broadcastLog(fmt.Sprintf("[CONTACT] New message from %s (%s)", newContact.Name, newContact.Email))

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func HandleProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		var projects []Project
		db.Find(&projects)
		json.NewEncoder(w).Encode(projects)
	} else {
		// Protected routes for admin
		claims, err := verifyJWT(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if claims.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if r.Method == http.MethodPost {
			var p Project
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			db.Save(&p)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		} else if r.Method == http.MethodDelete {
			id := r.URL.Query().Get("id")
			db.Delete(&Project{}, "id = ?", id)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func HandleContacts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims, err := verifyJWT(r)
	if err != nil || claims.Role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if r.Method == http.MethodGet {
		var contacts []Contact
		db.Find(&contacts)
		json.NewEncoder(w).Encode(contacts)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleLoginLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims, err := verifyJWT(r)
	if err != nil || claims.Role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if r.Method == http.MethodGet {
		var logs []LoginLog
		db.Find(&logs)
		json.NewEncoder(w).Encode(logs)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var logs []SystemLog
	if err := db.Order("id desc").Limit(30).Find(&logs).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reverse logs to be chronological (oldest to newest)
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	type LogResponse struct {
		Timestamp string `json:"timestamp"`
		Message   string `json:"message"`
	}

	res := make([]LogResponse, len(logs))
	for i, l := range logs {
		res[i] = LogResponse{
			Timestamp: l.CreatedAt.Format("15:04:05"),
			Message:   l.Message,
		}
	}

	json.NewEncoder(w).Encode(res)
}
