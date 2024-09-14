package server

import (
	"database/sql"
	"fmt"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"os"
)

type User struct {
	Id          string
	Name        string
	GoogleEmail string
}

type Session struct {
	Id        string
	ExpiresAt int64
}
type Workout struct {
	Weight int
	Sets   int
	Reps   int
}

type Exercise struct {
	Id       string
	Name     string
	Workouts []Workout
}

type Data struct {
	Name      string
	Exercises []Exercise
}

var data *sql.DB = nil

func CreateDb() {
	url := os.Getenv("TURSO_DATABASE_URL") + "?authToken=" + os.Getenv("TURSO_AUTH_TOKEN")
	db, err := sql.Open("libsql", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
	}
	data = db
	defer data.Close()
}

func GetDb() *sql.DB {
	return data
}
