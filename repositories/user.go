package repositories

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go-mini-test/models"
	"go-mini-test/sqlstore"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	db := sqlstore.DbConn()
	query, err := db.Query("SELECT id, name, age, address, phone FROM users ORDER BY id DESC")
	if err != nil {
		json.NewEncoder(w).Encode("Error " + err.Error())
		return
	}
	res := []models.User{}
	defer query.Close()

	for query.Next() {
		var user models.User
		err = query.Scan(&user.Id, &user.Name, &user.Age, &user.Address, &user.Phone)
		if err != nil {
			json.NewEncoder(w).Encode("Error " + err.Error())
			return
		}
		res = append(res, user)
	}
	json.NewEncoder(w).Encode(res)
}

func GetUserDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	db := sqlstore.DbConn()

	queryUser := "SELECT a.id, a.name, a.age, a.address, a.phone FROM users a WHERE a.id=?;"
	queryHobby := "SELECT userId, hobbyName FROM hobbies WHERE userId=?;"

	var user models.User
	var hobbies []models.Hobby

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
		var hobby models.Hobby
		if err := rows.Scan(&hobby.UserId, &hobby.HobbyName); err != nil {
			json.NewEncoder(w).Encode("Error " + err.Error())
			return
		}
		hobbies = append(hobbies, hobby)
	}

	user.Hobby = hobbies

	json.NewEncoder(w).Encode(user)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body)

	var user models.User
	json.Unmarshal(reqBody, &user)
	// json.NewEncoder(w).Encode(user)

	if user.Name == "" {
		json.NewEncoder(w).Encode("Error name is required")
		return
	}
	if user.Age == 0 {
		json.NewEncoder(w).Encode("Error age is required")
		return
	}
	if user.Address == "" {
		json.NewEncoder(w).Encode("Error address is required")
		return
	}
	if user.Phone == "" {
		json.NewEncoder(w).Encode("Error address is required")
		return
	}

	db := sqlstore.DbConn()

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

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	targetId := mux.Vars(r)["id"]
	reqBody, _ := io.ReadAll(r.Body)

	var user models.User
	json.Unmarshal(reqBody, &user)
	// json.NewEncoder(w).Encode(user)

	if user.Name == "" {
		json.NewEncoder(w).Encode("Error name is required")
		return
	}
	if user.Age == 0 {
		json.NewEncoder(w).Encode("Error age is required")
		return
	}
	if user.Address == "" {
		json.NewEncoder(w).Encode("Error address is required")
		return
	}
	if user.Phone == "" {
		json.NewEncoder(w).Encode("Error address is required")
		return
	}

	db := sqlstore.DbConn()

	// check if id exists
	var userCheck models.User
	if err := db.QueryRow(`SELECT * FROM users WHERE id=?`, targetId).Scan(&userCheck.Id, &userCheck.Name, &userCheck.Age, &userCheck.Address, &userCheck.Phone); err != nil {
		json.NewEncoder(w).Encode("Error user tidak ditemukan")
		return
	}

	// update user
	_, err := db.Exec(`UPDATE users SET name=?,age=?,address=?,phone=? WHERE id=?`, user.Name, user.Age, user.Address, user.Phone, targetId)
	if err != nil {
		json.NewEncoder(w).Encode("Error " + err.Error())
		return
	}

	// --delete hobby not in request (dissapear)
	var hobbiesDb []models.Hobby
	queryHobby := "SELECT userId, hobbyName FROM hobbies WHERE userId=?;"

	rows, err := db.Query(queryHobby, targetId)
	if err != nil {
		json.NewEncoder(w).Encode("Error " + err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var hobby models.Hobby
		if err := rows.Scan(&hobby.UserId, &hobby.HobbyName); err != nil {
			json.NewEncoder(w).Encode("Error " + err.Error())
			return
		}
		hobbiesDb = append(hobbiesDb, hobby)
	}

	// to list strings
	listId := []string{}

	for _, hobby := range user.Hobby {
		if hobby.Id != "" {
			listId = append(listId, hobby.Id)
		}
	}
	// fmt.Println(listId)

	var nonNilArgs []interface{}
	nonNilArgs = append(nonNilArgs, targetId)

	for _, id := range listId {
		if id != "" {
			nonNilArgs = append(nonNilArgs, id)
		}
	}
	// fmt.Println(nonNilArgs)

	args := nonNilArgs
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

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	db := sqlstore.DbConn()

	_, err := db.Exec(`DELETE FROM users WHERE id=?`, id)
	if err != nil {
		json.NewEncoder(w).Encode("Error " + err.Error())
		return
	}
	json.NewEncoder(w).Encode("User dihapus")
}
