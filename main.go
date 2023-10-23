package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Age     int64   `json:"age"`
	Address string  `json:"address"`
	Phone   string  `json:"phone"`
	Hobby   []Hobby `json:"hobby"`
}

type Hobby struct {
	Id        string `json:"id"`
	UserId    string `json:"userId"`
	HobbyName string `json:"hobbyName"`
}

type UserLogin struct {
	Id       string `json:"id"`
	UserId   string `json:"userId"`
	Username string `json:"username"`
	Password []byte `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var Users []User
var Hobbies []Hobby
var Login []UserLogin

func main() {

	db := dbConn()

	dropTable(db)
	createTable(db)

	// dummy data
	Users = []User{
		{Id: "1", Name: "Fatah At-Thariq", Age: 22, Address: "Cianjur", Phone: "087889550578"},
	}
	Hobbies = []Hobby{
		{Id: "1", UserId: "1", HobbyName: "Main game"},
		{Id: "2", UserId: "1", HobbyName: "Nonton anime"},
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("fatah123321"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
		return
	}
	Login = []UserLogin{
		{Id: "1", UserId: "1", Username: "fatah123", Password: hashedPassword},
	}

	// if no data then insert dummy
	count, err := getCount(db)
	if count == 0 && err == nil {
		for _, data := range Users {
			_, err := db.Exec(`INSERT INTO users(name, age, address, phone) VALUES (?,?,?,?)`, data.Name, data.Age, data.Address, data.Phone)
			if err != nil {
				log.Fatal(err)
			}
		}
		for _, data := range Hobbies {
			_, err := db.Exec(`INSERT INTO hobbies(userId, hobbyName) VALUES (?,?)`, data.UserId, data.HobbyName)
			if err != nil {
				log.Fatal(err)
			}
		}
		for _, data := range Login {
			_, err := db.Exec(`INSERT INTO user_logins(userId, username, password) VALUES (?,?,?)`, data.UserId, data.Username, data.Password)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	fmt.Println("started on port 10000")

	handleRequest()
}

func handleRequest() {
	// acts as route

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/api/login", login).Methods("POST")
	myRouter.HandleFunc("/api/user", createUser).Methods("POST")        // create
	myRouter.HandleFunc("/api/user", getUsers)                          // all
	myRouter.HandleFunc("/api/user/{id}", updateUser).Methods("PATCH")  // update
	myRouter.HandleFunc("/api/user/{id}", deleteUser).Methods("DELETE") // delete
	myRouter.HandleFunc("/api/user/{id}", getUserDetail)                // detail

	// replace nil with the new router
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := ""
	dbName := "go-mini-test"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

func dropTable(db *sql.DB) {
	if _, err := db.Exec("DROP TABLE IF EXISTS `hobbies`"); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec("DROP TABLE IF EXISTS `user_logins`"); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec("DROP TABLE IF EXISTS `users`"); err != nil {
		log.Fatal(err)
	}
}

func createTable(db *sql.DB) {
	// mysql
	query := `
		CREATE TABLE users(
			id INT AUTO_INCREMENT,
			name VARCHAR(255),
			age INT,
			address TEXT,
			phone VARCHAR(13),
			PRIMARY KEY(id)
		);
	`
	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}

	query = `
		CREATE TABLE hobbies (
			id INT AUTO_INCREMENT,
			userId INT,
			hobbyName TEXT,
			PRIMARY KEY(id),
			FOREIGN KEY (userId) REFERENCES users(id)
			ON DELETE CASCADE
		);
	`
	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}

	query = `
		CREATE TABLE user_logins (
			id INT AUTO_INCREMENT,
			userId INT,
			username VARCHAR(255),
			password VARCHAR(255),
			PRIMARY KEY(id),
			FOREIGN KEY (userId) REFERENCES users(id)
			ON DELETE CASCADE
		);
	`
	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}
}

func getCount(db *sql.DB) (count int, err error) {
	query := "SELECT COUNT(*) as count FROM users"
	row := db.QueryRow(query)

	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func createToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 12).Unix() // Token expiration time

	// Sign the token with a secret key
	tokenString, err := token.SignedString([]byte("this-is-secret"))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func login(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	reqBody, _ := io.ReadAll(r.Body)
	var userCred LoginRequest
	json.Unmarshal(reqBody, &userCred)

	username := userCred.Username
	password := userCred.Password

	var storedPassword string
	err := db.QueryRow("SELECT password FROM user_logins WHERE username = ?", username).Scan(&storedPassword)
	if err != nil {
		// fmt.Println(err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, errToken := createToken(username)
	if errToken != nil {
		http.Error(w, "Token creation failed", http.StatusUnauthorized)
		return
	}

	// Authentication successful
	json.NewEncoder(w).Encode(token)

}

func getUsers(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	query, err := db.Query("SELECT id, name, age, address, phone FROM users ORDER BY id DESC")
	if err != nil {
		json.NewEncoder(w).Encode("Error " + err.Error())
		return
	}
	res := []User{}
	defer query.Close()

	for query.Next() {
		var user User
		err = query.Scan(&user.Id, &user.Name, &user.Age, &user.Address, &user.Phone)
		if err != nil {
			json.NewEncoder(w).Encode("Error " + err.Error())
			return
		}
		res = append(res, user)
	}
	json.NewEncoder(w).Encode(res)
}

func getUserDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	db := dbConn()

	queryUser := "SELECT a.id, a.name, a.age, a.address, a.phone FROM users a WHERE a.id=?;"
	queryHobby := "SELECT userId, hobbyName FROM hobbies WHERE userId=?;"

	var user User
	var hobbies []Hobby

	// get user
	if err := db.QueryRow(queryUser, id).Scan(&user.Id, &user.Name, &user.Age, &user.Address, &user.Phone); err != nil {
		json.NewEncoder(w).Encode("Error " + err.Error())
		return
	}

	// get hobby
	rows, err := db.Query(queryHobby, id)
	if err != nil {
		json.NewEncoder(w).Encode("Error " + err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var hobby Hobby
		if err := rows.Scan(&hobby.UserId, &hobby.HobbyName); err != nil {
			json.NewEncoder(w).Encode("Error " + err.Error())
			return
		}
		hobbies = append(hobbies, hobby)
	}

	user.Hobby = hobbies

	json.NewEncoder(w).Encode(user)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body)

	var user User
	json.Unmarshal(reqBody, &user)
	// json.NewEncoder(w).Encode(user)

	db := dbConn()

	// insert user
	res, err := db.Exec(`INSERT INTO users(name, age, address, phone) VALUES (?,?,?,?)`, user.Name, user.Age, user.Address, user.Phone)
	if err != nil {
		json.NewEncoder(w).Encode("Error " + err.Error())
		return
	}
	userId, err := res.LastInsertId()

	// insert hobby
	for _, data := range user.Hobby {
		_, err := db.Exec(`INSERT INTO hobbies(userId, hobbyName) VALUES (?,?)`, userId, data.HobbyName)
		if err != nil {
			json.NewEncoder(w).Encode("Error " + err.Error())
			return
		}
	}

	json.NewEncoder(w).Encode("Data berhasil ditambahkan")
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	targetId := mux.Vars(r)["id"]
	reqBody, _ := io.ReadAll(r.Body)

	var user User
	json.Unmarshal(reqBody, &user)
	// json.NewEncoder(w).Encode(user)

	db := dbConn()

	// update user
	_, err := db.Exec(`UPDATE users SET name=?,age=?,address=?,phone=? WHERE id=?`, user.Name, user.Age, user.Address, user.Phone, targetId)
	if err != nil {
		json.NewEncoder(w).Encode("Error " + err.Error())
		return
	}

	// --delete hobby not in request (dissapear)
	var hobbiesDb []Hobby
	queryHobby := "SELECT userId, hobbyName FROM hobbies WHERE userId=?;"

	rows, err := db.Query(queryHobby, targetId)
	if err != nil {
		json.NewEncoder(w).Encode("Error " + err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var hobby Hobby
		if err := rows.Scan(&hobby.UserId, &hobby.HobbyName); err != nil {
			json.NewEncoder(w).Encode("Error " + err.Error())
			return
		}
		hobbiesDb = append(hobbiesDb, hobby)
	}

	// to map
	listId := []string{}

	for _, hobby := range user.Hobby {
		if hobby.Id != "" {
			listId = append(listId, hobby.Id)
		}
	}

	var nonNilArgs []interface{}
	nonNilArgs = append(nonNilArgs, targetId)

	for _, id := range listId {
		if id != "" {
			nonNilArgs = append(nonNilArgs, id)
		}
	}

	args := nonNilArgs
	// Prepare the query with placeholders
	query := "DELETE FROM hobbies WHERE userId=? AND id NOT IN (?" + strings.Repeat(",?", len(args)-2) + ") "
	// fmt.Println(args)
	// fmt.Println(query)
	_, err = db.Exec(query, args...)
	if err != nil {
		fmt.Println(err)
		return
	}

	// --end delete hobby not in request (dissapear)

	// update hobby
	for _, data := range user.Hobby {
		if data.Id != "" {
			// update existing hobby
			_, err := db.Exec(`UPDATE hobbies SET hobbyName=? WHERE userId=? AND id=?`, data.HobbyName, targetId, data.Id)
			if err != nil {
				json.NewEncoder(w).Encode("Error " + err.Error())
				return
			}
		} else {
			// insert new hobby
			_, err := db.Exec(`INSERT INTO hobbies(userId, hobbyName) VALUES (?,?)`, targetId, data.HobbyName)
			if err != nil {
				json.NewEncoder(w).Encode("Error " + err.Error())
				return
			}
		}
	}
	// -end update hobby

	json.NewEncoder(w).Encode("Data berhasil diubah")
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	db := dbConn()

	_, err := db.Exec(`DELETE FROM users WHERE id=?`, id)
	if err != nil {
		json.NewEncoder(w).Encode("Error " + err.Error())
		return
	}
	json.NewEncoder(w).Encode("User dihapus")
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the token's signing method and return the secret key
			return []byte("this-is-secret"), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if token.Valid {
			// Token is valid, continue with the next handler
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})
}
