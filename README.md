# Strava Adventure Game

As a passionate long-distance runner, I've always wanted to make something like this. The only thing that stopped me from making it was a lack of motivation (and free time ofc). But then, I discovered golang, and I saw its power and expressiveness. When I saw how quick and easy web development is with this language, I knew I found the right one... The language that offers power and simplicity, a very broad standard library, as well as a whole world of third-party libraries. But then I saw it actually has a full production-ready Web server within it, not just a simple development server like in Django. That totally blew me away. I’ve just found a new favorite language for my "Web ideas", as well as the motivation I was looking for.

## Intro

This relatively simple (for now) program pulls in activities from Strava — one of the most popular platforms for athletes — and calculates some basic stats (for now, just distance). That distance is then used in a fun way: the user picks a starting point and an ending point from a list of predefined locations. As they log more activities and their total distance increases, the app tracks how far they've "traveled" from the starting point toward the destination, showing their current virtual location. In other words, the user picks their own "adventure" — and that's where the app gets its name. The goal is to motivate athletes (and future athletes!) to keep moving by gamifying the process and making it fun. The data is shown on the main panel, and it's also embedded in each activity's description on Strava.

## What this application demonstrates

* Uses the database/sql package (I avoided third-party ORM modules because I wanted to get familiar with the standard database package first).
* Uses the net/http and html/template packages (bare net/http, with no web frameworks involved).
* Integrates with external REST APIs, such as Strava and OpenRouteService, for data retrieval.
* Delegates user authentication to an external service (Strava).
* Handles basic I/O operations and JSON encoding/decoding.
* Implements a basic session manager, using a simple in-memory map and session cookies.

## Setup & Run

* Build the binary in the standard way (install dependencies and run the build).
* Prepare the SQLite database: create a database using the provided schema file app.db.sql.
* Configure the application: rename the included config template config.ini.example to config.ini, and update the necessary fields.
* Run the binary.

## What has to be done

* ~~Strava bearer token refresh logic.~~
* Strava webhook handling.
* Determining the current location based on total distance and the calculated route between two places, with reverse geolocation used to look up the location name.
* Embedding calculated data into Strava activity descriptions.
* Detecting completed adventures.
* Logging.

## Plans for the future

* Robust config validation.
* Debug mode support: expose internal errors only in debug mode; otherwise, log them and return simplified, user-friendly messages.
* Database abstraction: consider introducing an ORM instead of using bare database/sql, or build a custom solution to auto-generate schemas, methods, and handle migrations.
* Map integration: display routes and current locations on an actual map instead of plain text.
* Unit testing.