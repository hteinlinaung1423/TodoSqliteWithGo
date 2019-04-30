package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"net/http"
	"strconv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/gorilla/mux"
)

type Todo struct {
	ID   int64  `json:"id"`
	IsCompleted bool `json:"isCompleted"`
	Text string `json:"text"`
}

type Todos []Todo

var mainDB *sql.DB

func main() {

	db, errOpenDB := sql.Open("sqlite3", "Todo.db")
	if errOpenDB != nil {
        log.Fatal("Error creating connection pool: " + errOpenDB.Error())
    }
    log.Printf("Connected!\n")
	mainDB = db


	r := mux.NewRouter()
    // Routes consist of a path and a handler function.
	r.HandleFunc("/", helloWorld)
	r.HandleFunc("/todos", getAll).Methods("GET")
	r.HandleFunc("/todos/{id}", getByID)
	r.HandleFunc("/todos", insert).Methods("POST")
	r.HandleFunc("/todos/update/{id}", updateByID).Methods("PUT")
	r.HandleFunc("/todos/delete/{id}", deleteByID).Methods("DELETE")
	// Bind to a port and pass our router in
	port := "8000"
	if os.Getenv("ASPNETCORE_PORT") != "" { // get enviroment variable that set by ACNM 
		port = os.Getenv("ASPNETCORE_PORT")
	}
	sport:=":"+port
    log.Fatal(http.ListenAndServe(sport, r))
	
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func getAll(w http.ResponseWriter, r *http.Request) {
	rows, err := mainDB.Query("SELECT * FROM Todos")
	checkErr(err)
	var todos Todos
	for rows.Next() {
		var todo Todo
		err = rows.Scan(&todo.ID, &todo.IsCompleted, &todo.Text)
		checkErr(err)
		todos = append(todos, todo)
	}
	jsonB, errMarshal := json.Marshal(todos)
	checkErr(errMarshal)
	fmt.Printf("Result : %s", string(jsonB))
	fmt.Fprintf(w, "%s", string(jsonB))
}

func getByID(w http.ResponseWriter, r *http.Request) {
	//id := r.URL.Query().Get("id")
	vars := mux.Vars(r)
	id := vars["id"]
	stmt, err := mainDB.Prepare(" SELECT * FROM todos where id = ?")
	checkErr(err)
	rows, errQuery := stmt.Query(id)
	checkErr(errQuery)
	var todo Todo
	for rows.Next() {
		err = rows.Scan(&todo.ID, &todo.IsCompleted, &todo.Text)
		checkErr(err)
	}
	jsonB, errMarshal := json.Marshal(todo)
	checkErr(errMarshal)
	fmt.Fprintf(w, "%s", string(jsonB))
}

func insert(w http.ResponseWriter, r *http.Request) {
	text := r.FormValue("Text")
	var todo Todo
	todo.Text = text
	stmt, err := mainDB.Prepare("INSERT INTO todos(Text,IsCompleted) values (?,?)")
	checkErr(err)
	result, errExec := stmt.Exec(todo.Text,true)
	checkErr(errExec)
	newID, errLast := result.LastInsertId()
	checkErr(errLast)
	todo.ID = newID
	jsonB, errMarshal := json.Marshal(todo)
	checkErr(errMarshal)
	fmt.Fprintf(w, "%s", string(jsonB))
}

func updateByID(w http.ResponseWriter, r *http.Request) {
	text := r.FormValue("Text")
	isCompleted := r.FormValue("IsCompleted")
	//id := r.URL.Query().Get(":id")
	vars := mux.Vars(r)
	id := vars["id"]
	var todo Todo
	ID, _ := strconv.ParseInt(id, 10, 0)
	IsCompleted,_:=strconv.ParseBool(isCompleted)
	todo.ID = ID
	todo.Text = text
	todo.IsCompleted=IsCompleted
	stmt, err := mainDB.Prepare("UPDATE todos SET Text = ?, IsCompleted=? WHERE id = ?")
	checkErr(err)
	result, errExec := stmt.Exec(todo.Text, todo.IsCompleted,todo.ID)
	checkErr(errExec)
	rowAffected, errLast := result.RowsAffected()
	checkErr(errLast)
	if rowAffected > 0 {
		jsonB, errMarshal := json.Marshal(todo)
		checkErr(errMarshal)
		fmt.Fprintf(w, "%s", string(jsonB))
	} else {
		fmt.Fprintf(w, "{row_affected=%d}", rowAffected)
	}

}

func deleteByID(w http.ResponseWriter, r *http.Request) {
	//id := r.URL.Query().Get(":id")
	vars := mux.Vars(r)
	id := vars["id"]
	stmt, err := mainDB.Prepare("DELETE FROM todos WHERE id = ?")
	checkErr(err)
	result, errExec := stmt.Exec(id)
	checkErr(errExec)
	rowAffected, errRow := result.RowsAffected()
	checkErr(errRow)
	fmt.Fprintf(w, "{row_affected=%d}", rowAffected)
}


func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
