CREATE TABLE user (
	id TEXT NOT NULL PRIMARY KEY,
	google_email TEXT NOT NULL,
	google_picture TEXT,
	name TEXT,
);

CREATE TABLE session (
	id TEXT NOT NULL PRIMARY KEY,
	user_id TEXT NOT NULL,
	expires_at INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE exercise (
	id TEXT NOT NULL PRIMARY KEY,
	user_id TEXT NOT NULL,
	name TEXT NOT NULL,
	FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE workout (
	id TEXT NOT NULL PRIMARY KEY,
	user_id TEXT NOT NULL,
	exercise_id TEXT NOT NULL,
	weight INTEGER NOT NULL,
	sets INTEGER NOT NULL,
	reps INTEGER NOT NULL,
	time INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES user(id)
	FOREIGN KEY (exercise_id) REFERENCES exercise(id)
);
