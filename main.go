// Package main implements RESTful service for solving 2 linear equations with 2 variables

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"app/equationSolver"

	"github.com/gorilla/mux"
)

// GetCalculation handles the calculation request
func GetCalculation(w http.ResponseWriter, r *http.Request) {
	equation1 := new(equationSolver.Equation)
	equation2 := new(equationSolver.Equation)
	log.Printf("Equation1: %s   Equation2: %s\n", r.FormValue("equation1"), r.FormValue("equation2"))

	// Init both equations
	initErr1 := equation1.Init(r.FormValue("equation1"))
	initErr2 := equation2.Init(r.FormValue("equation2"))

	if initErr1 != nil || initErr2 != nil {
		// Detect invalid input
		w.WriteHeader(http.StatusBadRequest)
		if initErr1 != nil {
			fmt.Fprint(w, initErr1)
			log.Println(initErr1)
		} else {
			fmt.Fprint(w, initErr2)
			log.Println(initErr2)
		}

	} else {
		// Solve the problem
		solution, calErr := equationSolver.Solve(equation1, equation2)
		if calErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, calErr)
			log.Println(calErr)
		} else {
			json.NewEncoder(w).Encode(solution)
			log.Printf("Solution: %c = %f   %c = %f\n", solution[0].Variable, solution[0].Value, solution[1].Variable, solution[1].Value)
		}
	}
}

// Hello handles main page request
func Hello(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	// Open log file for logging
	logFileName := "app.log"
	logFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Cannot opening file: %s", logFileName)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// RESTful service
	router := mux.NewRouter()
	router.HandleFunc("/", Hello)
	router.HandleFunc("/solution", GetCalculation).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}
