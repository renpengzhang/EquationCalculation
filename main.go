package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"unicode"

	"github.com/gorilla/mux"
)

// VariableAndCoef contains variable and its coefficient
type VariableAndCoef struct {
	variable rune
	coef     float64
}

// VariableAndValue contains variable and its value
type VariableAndValue struct {
	Variable rune    `json:"variable"`
	Value    float64 `json:"coefficient"`
}

// Equation contains information of an equation
type Equation struct {
	elements []VariableAndCoef
	value    float64
}

// Init equation obj by parsing the input equation string
func (equaiton *Equation) init(eString string) error {
	space := regexp.MustCompile(`\s+`)
	eString = space.ReplaceAllString(eString, "")

	if len(eString) == 0 {
		return errors.New("Empty String")
	}

	valueBeenSet := false
	beg := 0
	sign := 1.0
	for i := 0; i < len(eString); i++ {
		curChar := (rune)(eString[i])
		if unicode.IsLetter(curChar) {
			coefficient := sign
			if beg != i {
				f, err := strconv.ParseFloat(eString[beg:i], 64)
				if err != nil {
					return errors.New("Invalid String")
				}
				coefficient = coefficient * f
			}
			equaiton.elements = append(equaiton.elements, VariableAndCoef{curChar, coefficient})
		} else if curChar == '=' {
			f, err := strconv.ParseFloat(eString[i+1:len(eString)], 64)
			if err != nil {
				return errors.New("Invalid String")
			}
			equaiton.value = f
			valueBeenSet = true
		} else if curChar == '+' || curChar == '-' {
			if curChar == '+' {
				sign = 1
			} else {
				sign = -1
			}
			beg = i + 1
		}
	}

	if (len(equaiton.elements) != 2 && len(equaiton.elements) != 1) || !valueBeenSet {
		return errors.New("Invalid String")
	}

	if len(equaiton.elements) == 2 && (equaiton.elements[0].variable == equaiton.elements[1].variable) {
		return errors.New("Invalid String")
	}
	return nil
}

// Solve the equation
func calculate(equaiton1 *Equation, equaiton2 *Equation) ([]VariableAndValue, error) {
	if len(equaiton1.elements) == 1 && len(equaiton2.elements) == 1 {
		if equaiton1.elements[0].variable == equaiton2.elements[0].variable {
			return nil, errors.New("Not Match Equation")
		} else {
			equaiton1.elements = append(equaiton1.elements, VariableAndCoef{equaiton2.elements[0].variable, 0})
			equaiton2.elements = append(equaiton2.elements, VariableAndCoef{equaiton1.elements[0].variable, 0})
		}
	} else if len(equaiton1.elements) == 1 {
		if equaiton1.elements[0].variable == equaiton2.elements[0].variable {
			equaiton1.elements = append(equaiton1.elements, VariableAndCoef{equaiton2.elements[1].variable, 0})
		} else {
			equaiton1.elements = append(equaiton1.elements, VariableAndCoef{equaiton2.elements[0].variable, 0})
		}
	} else if len(equaiton2.elements) == 1 {
		if equaiton2.elements[0].variable == equaiton1.elements[0].variable {
			equaiton2.elements = append(equaiton2.elements, VariableAndCoef{equaiton1.elements[1].variable, 0})
		} else {
			equaiton2.elements = append(equaiton2.elements, VariableAndCoef{equaiton1.elements[0].variable, 0})
		}
	}

	sort.Slice(equaiton1.elements, func(i, j int) bool {
		return equaiton1.elements[i].variable < equaiton1.elements[j].variable
	})

	sort.Slice(equaiton2.elements, func(i, j int) bool {
		return equaiton2.elements[i].variable < equaiton2.elements[j].variable
	})

	VC11 := equaiton1.elements[0]
	VC12 := equaiton1.elements[1]
	VC21 := equaiton2.elements[0]
	VC22 := equaiton2.elements[1]
	value1 := equaiton1.value
	value2 := equaiton2.value

	if (VC11.variable != VC21.variable) || (VC12.variable != VC22.variable) {
		return nil, errors.New("Not Match Equation")
	}

	if math.Abs(VC11.coef*VC22.coef-VC12.coef*VC21.coef) < 0.0000001 {
		if math.Abs(value1*VC21.coef-value2*VC11.coef) < 0.0000001 {
			return nil, errors.New("Infinite Number of Solutions")
		} else {
			return nil, errors.New("No Solutions")
		}
	}

	valueForVariable1 := (value1*VC22.coef - value2*VC12.coef) / (VC11.coef*VC22.coef - VC12.coef*VC21.coef)
	valueForVariable2 := (value2*VC11.coef - value1*VC21.coef) / (VC11.coef*VC22.coef - VC12.coef*VC21.coef)

	var solution []VariableAndValue
	solution = append(solution, VariableAndValue{VC11.variable, valueForVariable1})
	solution = append(solution, VariableAndValue{VC12.variable, valueForVariable2})
	return solution, nil
}

// GetCalculation handles the calculation request
func GetCalculation(w http.ResponseWriter, r *http.Request) {
	equation1 := new(Equation)
	equation2 := new(Equation)
	log.Printf("Equation1: %s   Equation2: %s\n", r.FormValue("equation1"), r.FormValue("equation2"))

	initErr1 := equation1.init(r.FormValue("equation1"))
	initErr2 := equation2.init(r.FormValue("equation2"))
	if initErr1 != nil || initErr2 != nil {
		w.WriteHeader(http.StatusBadRequest)
		if initErr1 != nil {
			fmt.Fprint(w, initErr1)
			log.Println(initErr1)
		} else {
			fmt.Fprint(w, initErr2)
			log.Println(initErr2)
		}

	} else {
		solution, calErr := calculate(equation1, equation2)
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

// hello handles main page request
func hello(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	logFileName := "app.log"
	logFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("Cannot opening file: %s", logFileName)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	router := mux.NewRouter()
	router.HandleFunc("/", hello)
	router.HandleFunc("/calculateEquation", GetCalculation).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
	// test
	// comment
}
