# Gym Journal

This is gym journal. An open source tracking web application inspired by Stupid Simple Macro Tracker

### Goals
* Provide a simple user interface to see trends in training data
* Implement a web application using Go and HTMX mainly using the standard library

### Technologies
* Go
* HTMX
* Google Oauth (Goth)
* Turso

### Self Hosting
Environment variables that need to be set:
```
GOOGLE_CLIENT_ID
GOOGLE_SECRET
TURSO_DATABASE_URL
TURSO_AUTH_TOKEN
```
The database schema can be found in the src/sql folder
