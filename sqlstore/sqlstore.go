package sqlstore

import (
	"database/sql"
	"log"

	"go-mini-test/models"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var Users []models.User
var Hobbies []models.Hobby
var Login []models.UserLogin

func DbConn() (db *sql.DB) {
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

func DropTable(db *sql.DB) {
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

func CreateTable(db *sql.DB) {
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

func CreateDummy(db *sql.DB) {
	// dummy data
	Users = []models.User{
		{Id: "1", Name: "Fatah At-Thariq", Age: 22, Address: "Cianjur", Phone: "087889550578"},
	}
	Hobbies = []models.Hobby{
		{Id: "1", UserId: "1", HobbyName: "Main game"},
		{Id: "2", UserId: "1", HobbyName: "Nonton anime"},
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("fatah123321"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
		return
	}
	Login = []models.UserLogin{
		{Id: "1", UserId: "1", Username: "fatah123", Password: hashedPassword},
	}

	// if no data then insert dummy
	count, err := GetCount(db)
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
}

func GetCount(db *sql.DB) (count int, err error) {
	query := "SELECT COUNT(*) as count FROM users"
	row := db.QueryRow(query)

	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
