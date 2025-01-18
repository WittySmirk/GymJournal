package server

import (
	"database/sql"
	"fmt"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"os"
)

type User struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	GoogleEmail string `json:"googleEmail"`
}

type Session struct {
	Id        string `json:"id"`
	ExpiresAt int64  `json:"expiresAt"`
}
type Workout struct {
	Weight int `json:"weight"`
	Sets   int `json:"sets"`
	Reps   int `json:"reps"`
}

type Exercise struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Workouts []Workout `json:"workouts"`
}

type Data struct {
	NameNotValid bool       `json:"nameNotValid"`
	Name         string     `json:"name"`
	Exercises    []Exercise `json:"exercises"`
}

var data *sql.DB = nil

func CreateDb() {
	url := os.Getenv("TURSO_DATABASE_URL") + "?authToken=" + os.Getenv("TURSO_AUTH_TOKEN")
	db, err := sql.Open("libsql", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
	}
	data = db
}

func GetDb() *sql.DB {
	return data
}

func CloseDb() {
	data.Close()
}
