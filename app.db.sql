BEGIN TRANSACTION;
DROP TABLE IF EXISTS "Adventure";
CREATE TABLE IF NOT EXISTS "Adventure" (
	"athlete_id"	INTEGER NOT NULL,
	"start_location"	INTEGER NOT NULL,
	"end_location"	INTEGER NOT NULL,
	"current_location_lat"	REAL NOT NULL,
	"current_location_lon"	REAL NOT NULL,
	"current_location_index_on_route"	INTEGER NOT NULL,
	"current_location_name"	TEXT NOT NULL,
	"current_distance"	REAL NOT NULL DEFAULT 0,
	"total_distance"	REAL NOT NULL,
	"completed"	INTEGER NOT NULL DEFAULT 0,
	"start_date"	INTEGER NOT NULL,
	"end_date"	INTEGER,
	PRIMARY KEY("athlete_id","start_location","end_location"),
	FOREIGN KEY("end_location") REFERENCES "Location"("id") ON DELETE CASCADE,
	FOREIGN KEY("start_location") REFERENCES "Location"("id") ON DELETE CASCADE,
	FOREIGN KEY("athlete_id") REFERENCES "Athlete"("id") ON DELETE CASCADE
);
COMMIT;
