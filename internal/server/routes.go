package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	//	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/markbates/goth/gothic"
	"net/http"
	//	"strconv"
	//	"strings"
	"time"
)

func CorsMiddleWare(n http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required
		n(w, r)
	}
}

// TODO: Reimplement this to get and post user data for real
func GetUserFromContext(r *http.Request) *User {
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

func CheckSession(shouldexist bool, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := GetDb()
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
		http.Redirect(w, r, "http://localhost:5173/app", http.StatusSeeOther)
	}
}

// TODO: Refactor to react and not htmx
func CreateRoutes() http.Handler {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("public"))
	mux.Handle("/public/", http.StripPrefix("/public/", fs))

	//mux.HandleFunc("/", CheckSession(false, index))
	mux.HandleFunc("/auth/google/callback", CallbackHandle)
	//mux.HandleFunc("/logout/google", LogoutHandle)
	mux.HandleFunc("/auth/google", CheckSession(false, AuthHandle))
	mux.HandleFunc("GET /app", CorsMiddleWare(appGet))
	mux.HandleFunc("POST /app", CorsMiddleWare(appPost))

	return mux
}

/*
func index(w http.ResponseWriter, r *http.Request) {
	templ, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templ.Execute(w, nil)
}
*/

func CallbackHandle(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	q.Add("provider", "google")
	r.URL.RawQuery = q.Encode()
	tuser, gerr := gothic.CompleteUserAuth(w, r)
	if gerr != nil {
		fmt.Fprintln(w, gerr)
		return
	}

	// Check if user exists in db
	var myuser User
	db := GetDb()
	dUser := db.QueryRow("SELECT id, google_email FROM user WHERE google_email = ?", tuser.Email)
	dberr := dUser.Scan(&myuser.Id, &myuser.GoogleEmail) // See if query can fit into a User if not make an error

	if dberr != nil {
		if dberr == sql.ErrNoRows {
			// User does not exist error
			userId := uuid.New()
			_, uerr := db.Exec("INSERT INTO user (id, google_email) VALUES (?, ?)", userId, tuser.Email)

			if uerr != nil {
				fmt.Fprintln(w, uerr)
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
		fmt.Fprintln(w, serr)
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
	http.Redirect(w, r, "http://localhost:5173/app", http.StatusTemporaryRedirect)
}

/*
func LogoutHandle(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
*/

func AuthHandle(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	q.Add("provider", "google")
	r.URL.RawQuery = q.Encode()
	if _, err := gothic.CompleteUserAuth(w, r); err == nil {
		http.Redirect(w, r, "http://localhost:5173/app", http.StatusTemporaryRedirect)
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func appGet(w http.ResponseWriter, r *http.Request) {
	// TODO: maybe this works i dont remember/know
	/*
			myuser := GetUserFromContext(r)
			if myuser == nil {
				fmt.Println("error context was nil")
				return
			}


		db := GetDb()
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
	*/

	data := Data{
		Name: "wyatt",
		Exercises: []Exercise{
			{
				Id:   "0",
				Name: "bench",
			},
			{
				Id:   "1",
				Name: "squat",
			},
		},
	}

	jsonbytes, jerr := json.Marshal(data)
	if jerr != nil {
		http.Error(w, jerr.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonbytes)
}

type Body struct {
	Create bool   `json:"create"`
	Name   string `json:"name"`
	Id     string `json:"id"`
	Weight int    `json:"weight"`
	Sets   int    `json:"sets"`
	Reps   int    `json:"reps"`
}

func appPost(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var body Body
	berr := decoder.Decode(&body)
	if berr != nil {
		http.Error(w, berr.Error(), http.StatusBadRequest)
	}
	// TODO: set up a way to check stuffs
	fmt.Println(body.Id, body.Name)

	/*
		myuser := GetUserFromContext(r)
		if myuser == nil {
			fmt.Println("error context was nil")
			return
		}

		validate := validator.New()
		name := r.FormValue("name")
		verr := validate.Var(name, "ascii")
		if verr != nil {
			http.Error(w, "name not accepted", http.StatusNotAcceptable)
			return
		}

		db := GetDb()
		_, derr := db.Exec("UPDATE user SET name = ? WHERE id = ?", name, myuser.Id)
		if derr != nil {
			fmt.Fprintln(w, derr.Error())
			return
		}
		http.Redirect(w, r, "/app", http.StatusSeeOther)
	*/
}
