package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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

// TODO: Maybe break into packages to make this more elegant

func checksession(db *sql.DB, shouldexist bool, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionId, err := r.Cookie("session_id")
		if err != nil {
			if err == http.ErrNoCookie {
				if !shouldexist {
					h(w, r)
					return
				}
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
			fmt.Fprintln(w, err.Error())
			return
		}
		if shouldexist {
			var mysession Session
			mysession.Id = sessionId.Value
			dSession := db.QueryRow("SELECT expires_at FROM session WHERE id = ?", mysession.Id)
			dSerr := dSession.Scan(&mysession.ExpiresAt)

			if dSerr != nil {
				fmt.Fprintln(w, dSerr.Error())
			}

			// Delete session if expired, update active session if halfway dead
			if mysession.ExpiresAt < time.Now().Unix() {
				// Session is expired
				_, derr := db.Exec("DELETE FROM session WHERE id = ?", mysession.Id)
				if derr != nil {
					fmt.Fprintln(w, derr.Error())
				}
				http.SetCookie(w, &http.Cookie{
					Name:     "session_id",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
				})
				http.Redirect(w, r, "/", http.StatusSeeOther)
			} else if mysession.ExpiresAt-time.Now().Unix() < (7 * 24 * 60 * 60) {
				newTime := time.Now().Unix() + (15 * 24 * 60 * 60)
				_, derr := db.Exec("UPDATE session SET expires_at = ? WHERE id = ?", newTime, mysession.Id)
				if derr != nil {
					fmt.Fprintln(w, derr.Error())
					return
				}
				http.SetCookie(w, &http.Cookie{
					Name:     "session_id",
					Value:    mysession.Id,
					Path:     "/",
					MaxAge:   15 * 24 * 60 * 60,
					HttpOnly: true,
				})
			}

			const userKey string = "user"
			// TODO: Might want to include name in the context if more templates than app eventually need it
			var myuser User
			dUser := db.QueryRow("SELECT user_id FROM session WHERE id = ?", mysession.Id)
			dUerr := dUser.Scan(&myuser.Id)
			if dUerr != nil {
				fmt.Fprintln(w, dUerr.Error())
				return
			}

			ctx := context.WithValue(r.Context(), userKey, myuser.Id)
			h(w, r.WithContext(ctx))
			return
		}
		http.Redirect(w, r, "/app", http.StatusSeeOther)
	}
}

func getUserFromContext(r *http.Request) *User {
	const userKey string = "user"
	userid, ok := r.Context().Value(userKey).(string)

	if !ok {
		fmt.Println("user id not found in context")
		return nil
	}

	user := User{
		Id: userid,
	}

	return &user
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("failed to load env")
		return
	}

	url := os.Getenv("TURSO_DATABASE_URL") + "?authToken=" + os.Getenv("TURSO_AUTH_TOKEN")

	db, err := sql.Open("libsql", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
	}

	defer db.Close()

	goth.UseProviders(google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_SECRET"), "http://localhost:8080/auth/google/callback"))

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", checksession(db, false, func(w http.ResponseWriter, r *http.Request) {
		templ, err := template.ParseFiles("templates/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templ.Execute(w, nil)
	}))

	// Routes dealing with authenitcation
	http.HandleFunc("/auth/google/callback", checksession(db, false, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Add("provider", "google")
		r.URL.RawQuery = q.Encode()
		tuser, gerr := gothic.CompleteUserAuth(w, r)
		if gerr != nil {
			fmt.Fprintln(w, err)
			return
		}

		// Check if user exists in db
		var myuser User
		dUser := db.QueryRow("SELECT id, google_email FROM user WHERE google_email = ?", tuser.Email)
		dberr := dUser.Scan(&myuser.Id, &myuser.GoogleEmail) // See if query can fit into a User if not make an error

		if dberr != nil {
			if dberr == sql.ErrNoRows {
				// User does not exist error
				userId := uuid.New()
				_, uerr := db.Exec("INSERT INTO user (id, google_email) VALUES (?, ?)", userId, tuser.Email)

				if uerr != nil {
					fmt.Fprintln(w, err)
					return
				}
				myuser.Id = userId.String()

				// Insert default stuff in
				_, eerr := db.Exec("INSERT INTO exercise (id, user_id, name) VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?)", uuid.New(), myuser.Id, "Bench", uuid.New(), myuser.Id, "Squat", uuid.New(), myuser.Id, "Deadlift")
				if eerr != nil {
					fmt.Fprintln(w, eerr)
					return
				}
			} else {
				// DB just failed or smthn
				fmt.Fprintln(w, dberr.Error())
				return
			}
		}

		// Create session
		fmt.Println(myuser.Id)
		sessionId := uuid.New()
		expiresAt := time.Now().Unix() + (15 * 24 * 60 * 60) // Make session expire in 15 days
		fmt.Println(expiresAt)
		_, serr := db.Exec("INSERT INTO session (id, user_id, expires_at) VALUES (?, ?, ?)", sessionId, myuser.Id, expiresAt)
		if serr != nil {
			fmt.Fprintln(w, err)
			return
		}

		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    sessionId.String(),
			Path:     "/",
			MaxAge:   15 * 24 * 60 * 60,
			HttpOnly: true,
			Secure:   false, // Change to true in production
		}

		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/app", http.StatusTemporaryRedirect)
	}))

	http.HandleFunc("/logout/google", func(w http.ResponseWriter, r *http.Request) {
		gothic.Logout(w, r)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})

	http.HandleFunc("/auth/google", checksession(db, false, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Add("provider", "google")
		r.URL.RawQuery = q.Encode()
		if _, err := gothic.CompleteUserAuth(w, r); err == nil {
			http.Redirect(w, r, "/app", http.StatusTemporaryRedirect)
		} else {
			gothic.BeginAuthHandler(w, r)
		}
	}))

	// Routes dealing with application
	http.HandleFunc("/app", checksession(db, true, func(w http.ResponseWriter, r *http.Request) {
		myuser := getUserFromContext(r)
		if myuser == nil {
			fmt.Println("error context was nil")
			return
		}

		if r.Method == "GET" {
			var name sql.NullString
			dUser := db.QueryRow("SELECT name FROM user WHERE id = ?", myuser.Id)
			dUerr := dUser.Scan(&name)
			if dUerr != nil {
				fmt.Fprintln(w, dUerr.Error())
				return
			}

			if !name.Valid {
				// If name is not valid ask for name
				templ, err := template.ParseFiles("templates/makename.html")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				templ.Execute(w, nil)
			}

			myuser.Name = name.String

			data := Data{
				Name: myuser.Name,
			}

			templ, err := template.ParseFiles("templates/app.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			templ.Execute(w, data)
		} else if r.Method == "POST" {
			validate := validator.New()
			name := r.FormValue("name")
			verr := validate.Var(name, "ascii")
			if verr != nil {
				http.Error(w, "name not accepted", http.StatusNotAcceptable)
				return
			}

			_, derr := db.Exec("UPDATE user SET name = ? WHERE id = ?", name, myuser.Id)
			if derr != nil {
				fmt.Fprintln(w, derr.Error())
				return
			}
			http.Redirect(w, r, "/app", http.StatusSeeOther)
		}
	}))

	http.HandleFunc("/app/modal/", checksession(db, true, func(w http.ResponseWriter, r *http.Request) {
		id := strings.Split(r.URL.Path, "/")[3]
		if r.Method == "GET" {
			templ, err := template.ParseFiles("templates/modal.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if id == "" {
				templ.Execute(w, nil)
			} else {
				data := Exercise{
					Id: id,
				}
				templ.Execute(w, data)
			}
		} else if r.Method == "POST" {
			validate := validator.New()

			myuser := getUserFromContext(r)

			if id == "" {
				name := r.FormValue("name")
				verr := validate.Var(name, "ascii")
				if verr != nil || len(name) == 0 {
					http.Error(w, "title not accepted", http.StatusNotAcceptable)
					return
				}
				_, derr := db.Exec("INSERT INTO exercise (id, user_id, name) VALUES (?, ?, ?)", uuid.New(), myuser.Id, name)
				if derr != nil {
					fmt.Fprintln(w, derr.Error())
					return
				}
			} else {
				weight := r.FormValue("weight")
				sets := r.FormValue("sets")
				reps := r.FormValue("reps")

				convWeight, werr := strconv.Atoi(weight)
				if werr != nil {
					http.Error(w, werr.Error(), http.StatusNotAcceptable)
					return
				}
				convSets, serr := strconv.Atoi(sets)
				if serr != nil {
					http.Error(w, serr.Error(), http.StatusNotAcceptable)
					return
				}
				convReps, rerr := strconv.Atoi(reps)
				if rerr != nil {
					http.Error(w, rerr.Error(), http.StatusNotAcceptable)
					return
				}
				_, derr := db.Exec("INSERT INTO workout (id, user_id, exercise_id, weight, sets, reps, time) VALUES (?, ?, ?, ?, ?, ?, ?)", uuid.New(), myuser.Id, id, convWeight, convSets, convReps, time.Now().Unix())
				if derr != nil {
					fmt.Fprintln(w, derr.Error())
					return
				}
			}
			http.Redirect(w, r, "/app", http.StatusSeeOther)
		}
	}))

	http.HandleFunc("/app/exercisebuttons", checksession(db, true, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			templ, terr := template.ParseFiles("templates/exercisebuttons.html")
			if terr != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			myuser := getUserFromContext(r)

			var exercises []Exercise
			dExercises, dEerr := db.Query("SELECT id, name FROM exercise WHERE user_id = ?", myuser.Id)
			if dEerr != nil {
				fmt.Fprintln(w, dEerr.Error())
				return
			}

			for dExercises.Next() {
				var ex Exercise
				err := dExercises.Scan(&ex.Id, &ex.Name)
				if err != nil {
					fmt.Fprintln(w, err.Error())
					return
				}
				exercises = append(exercises, ex)
			}
			data := Data{
				Exercises: exercises,
			}
			templ.Execute(w, data)
		}
	}))

	http.HandleFunc("/app/givememodal/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.Split(r.URL.Path, "/")[3]
		if r.Method == "GET" {
			if id == "" {
				templ, terr := template.ParseFiles("templates/modal_forms/new_exercise.html")
				if terr != nil {
					http.Error(w, terr.Error(), http.StatusInternalServerError)
					return
				}
				templ.Execute(w, nil)
			} else {
				templ, terr := template.ParseFiles("templates/modal_forms/exercise.html")
				if terr != nil {
					http.Error(w, terr.Error(), http.StatusInternalServerError)
					return
				}

				var exercise Exercise
				dExercise := db.QueryRow("SELECT name FROM exercise WHERE id = ?", id)
				dErr := dExercise.Scan(&exercise.Name)
				if dErr != nil {
					fmt.Fprintln(w, dErr.Error())
					return
				}

				exercise.Id = id
				templ.Execute(w, exercise)
			}
		}
	})

	http.ListenAndServe(":8080", nil)
}
