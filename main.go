package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "admin"
	dbname   = "go-test"
)

type User struct {
	Id       int
	Username string
	Email    string
	Age      int
}

var users map[int]User

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	userArray := []User{}
	for _, v := range users {
		userArray = append(userArray, v)
	}
	json.NewEncoder(w).Encode(userArray)
}

func getOneUser(wr http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userId, err := strconv.Atoi(params["id"])

	if err == nil {
		if _, ok := users[userId]; ok {
			json.NewEncoder(wr).Encode(users[userId])
		} else {
			wr.WriteHeader(http.StatusNotFound)
			wr.Write([]byte("User not found\n"))
		}
	} else {
		wr.WriteHeader(http.StatusInternalServerError)
		wr.Write([]byte("Cannot convert user id to integer value"))
	}
}

func createUser(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") == "application/json" {
		if r.Method == "POST" {
			body, err := ioutil.ReadAll(r.Body)

			if err == nil {
				var user User
				json.Unmarshal([]byte(body), &user)

				// check for incorrect json format
				if user.Username == "" || user.Email == "" || user.Age <= 0 || user.Id == 0 {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Please, supply user data in JSON format."))
					return
				}

				// check if user already exists
				// if it doesn't, create new user
				if _, ok := users[user.Id]; !ok {
					users[user.Id] = user
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - User has been successfully created:\n"))
					w.Write([]byte(
						strconv.Itoa(user.Id) + " - " +
							user.Username + " - " +
							user.Email + " - " +
							strconv.Itoa(user.Age)))
				} else {
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte("409 - User with this id already exists"))
				}
			} else {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Incorrect JSON format"))
			}
		}
	}
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") == "application/json" {
		params := mux.Vars(r)
		userId, _ := strconv.Atoi(params["id"])

		body, err := ioutil.ReadAll(r.Body)

		if err == nil {
			var user User
			json.Unmarshal([]byte(body), &user)

			// check for incorrect json format
			if user.Username == "" || user.Email == "" || user.Age <= 0 {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Please, supply user data in JSON format."))
				return
			}

			// if user doesn't exists, create a new one
			if _, ok := users[userId]; !ok {
				users[userId] = user
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("201 - User has been successfully created:\n"))
				w.Write([]byte(
					strconv.Itoa(user.Id) + " - " +
						user.Username + " - " +
						user.Email + " - " +
						strconv.Itoa(user.Age)))
			} else {
				// if user exists, update it
				users[userId] = user
				w.WriteHeader(http.StatusNoContent)
			}
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Incorrect JSON format"))
		}
	}
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userId, err := strconv.Atoi(params["id"])

	if err == nil {
		if _, ok := users[userId]; ok {
			delete(users, userId)
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("User with id \"" + params["id"] + "\" not found"))
		}
	}
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	users = make(map[int]User)

	const API_ROUTE = "/api/v1"

	router := mux.NewRouter()

	const USER_ROUTE = API_ROUTE + "/user"
	const USER_ROUTE_ID = USER_ROUTE + "/{id}"

	router.HandleFunc(USER_ROUTE, getAllUsers).Methods("GET")
	router.HandleFunc(USER_ROUTE, createUser).Methods("POST")
	router.HandleFunc(USER_ROUTE_ID, getOneUser).Methods("GET")
	router.HandleFunc(USER_ROUTE_ID, updateUser).Methods("PUT")
	router.HandleFunc(USER_ROUTE_ID, deleteUser).Methods("DELETE")

	const PORT = "3030"
	fmt.Println("Server started on port " + PORT)
	log.Fatal(http.ListenAndServe(":"+PORT, router))
}
