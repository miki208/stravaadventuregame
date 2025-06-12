BEGIN TRANSACTION;
DROP TABLE IF EXISTS "Athlete";
CREATE TABLE IF NOT EXISTS "Athlete" (
	"id"	INTEGER NOT NULL,
	"username"	TEXT NOT NULL,
	"first_name"	TEXT,
	"last_name"	TEXT,
	"city"	TEXT,
	"country"	TEXT,
	"sex"	TEXT,
	"is_admin"	INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("id")
);
DROP TABLE IF EXISTS "StravaCredentials";
CREATE TABLE IF NOT EXISTS "StravaCredentials" (
	"athlete_id"	INTEGER NOT NULL,
	"access_token"	TEXT NOT NULL,
	"refresh_token"	TEXT NOT NULL,
	"expires_at"	INTEGER NOT NULL,
	FOREIGN KEY("athlete_id") REFERENCES "Athlete"("id") ON DELETE CASCADE,
	PRIMARY KEY("athlete_id")
);
DROP TABLE IF EXISTS "Location";
CREATE TABLE IF NOT EXISTS "Location" (
	"id"	INTEGER NOT NULL,
	"lat"	REAL NOT NULL,
	"lon"	REAL NOT NULL,
	"name"	TEXT NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT)
);
DROP TABLE IF EXISTS "Adventure";
CREATE TABLE IF NOT EXISTS "Adventure" (
	"athlete_id"	INTEGER NOT NULL,
	"start_location"	INTEGER NOT NULL,
	"end_location"	INTEGER NOT NULL,
	"current_location_name"	TEXT NOT NULL,
	"current_distance"	REAL NOT NULL DEFAULT 0,
	"total_distance"	REAL NOT NULL,
	"completed"	INTEGER NOT NULL DEFAULT 0,
	FOREIGN KEY("end_location") REFERENCES "Location"("id") ON DELETE CASCADE,
	FOREIGN KEY("athlete_id") REFERENCES "Athlete"("id") ON DELETE CASCADE,
	FOREIGN KEY("start_location") REFERENCES "Location"("id") ON DELETE CASCADE,
	PRIMARY KEY("athlete_id","start_location","end_location")
);
COMMIT;
