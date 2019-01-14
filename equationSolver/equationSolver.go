// Package equationSolver implements util methods for solving 2 linear equations with 2 variables

package equationSolver

import (
	"errors"
	"math"
	"regexp"
	"sort"
	"strconv"
	"unicode"
)

// VariableAndCoef contains variable and its coefficient
type VariableAndCoef struct {
	Variable rune
	Coef     float64
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
func (equaiton *Equation) Init(eString string) error {
	// Remove extra spaces
	space := regexp.MustCompile(`\s+`)
	eString = space.ReplaceAllString(eString, "")

	// Check whether the input is valid equaiton string.
	// The valid form should be Ax +/- By = C or Az = C
	validEquation := regexp.MustCompile(`^[+-]?(([0-9]*[.])?[0-9]+)?[a-zA-Z]([+-](([0-9]*[.])?[0-9]+)?[a-zA-Z])?=([+-]?([0-9]*[.])?[0-9]+)$`)
	if !validEquation.MatchString(eString) {
		return errors.New("Invalid String")
	}

	valueBeenSet := false
	beg := 0
	sign := 1.0

	// Walk through the string to build Equation obj
	for i := 0; i < len(eString); i++ {
		curChar := (rune)(eString[i])

		if unicode.IsLetter(curChar) {
			// Found variable
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
			// Set right value of the equation
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

	// Invalid string if there is no or more than 3 varaiables, or no right value
	if (len(equaiton.elements) != 2 && len(equaiton.elements) != 1) || !valueBeenSet {
		return errors.New("Invalid String")
	}

	// Invalid string if 2 varibales share the same name
	if len(equaiton.elements) == 2 && (equaiton.elements[0].Variable == equaiton.elements[1].Variable) {
		return errors.New("Invalid String")
	}

	return nil
}

// Solve the equation
func Solve(equaiton1 *Equation, equaiton2 *Equation) ([]VariableAndValue, error) {
	if len(equaiton1.elements) == 1 && len(equaiton2.elements) == 1 {
		// Case that both equations have only 1 variable and needed to be completed
		if equaiton1.elements[0].Variable == equaiton2.elements[0].Variable {
			return nil, errors.New("Equations Not Match")
		} else {
			equaiton1.elements = append(equaiton1.elements, VariableAndCoef{equaiton2.elements[0].Variable, 0})
			equaiton2.elements = append(equaiton2.elements, VariableAndCoef{equaiton1.elements[0].Variable, 0})
		}
	} else if len(equaiton1.elements) == 1 {
		// Case that 1 equation has only 1 variable and needed to be completed
		if equaiton1.elements[0].Variable == equaiton2.elements[0].Variable {
			equaiton1.elements = append(equaiton1.elements, VariableAndCoef{equaiton2.elements[1].Variable, 0})
		} else {
			equaiton1.elements = append(equaiton1.elements, VariableAndCoef{equaiton2.elements[0].Variable, 0})
		}
	} else if len(equaiton2.elements) == 1 {
		// Case that 1 equation has only 1 variable and needed to be completed
		if equaiton2.elements[0].Variable == equaiton1.elements[0].Variable {
			equaiton2.elements = append(equaiton2.elements, VariableAndCoef{equaiton1.elements[1].Variable, 0})
		} else {
			equaiton2.elements = append(equaiton2.elements, VariableAndCoef{equaiton1.elements[0].Variable, 0})
		}
	}

	// Sort the variables for both equations
	sort.Slice(equaiton1.elements, func(i, j int) bool {
		return equaiton1.elements[i].Variable < equaiton1.elements[j].Variable
	})

	sort.Slice(equaiton2.elements, func(i, j int) bool {
		return equaiton2.elements[i].Variable < equaiton2.elements[j].Variable
	})

	VC11 := equaiton1.elements[0]
	VC12 := equaiton1.elements[1]
	VC21 := equaiton2.elements[0]
	VC22 := equaiton2.elements[1]
	value1 := equaiton1.value
	value2 := equaiton2.value

	// Case that the variables' names are not matched
	if (VC11.Variable != VC21.Variable) || (VC12.Variable != VC22.Variable) {
		return nil, errors.New("Equations Not Match")
	}

	// Case that 2 equations are relative
	if math.Abs(VC11.Coef*VC22.Coef-VC12.Coef*VC21.Coef) < 0.0000001 {
		if math.Abs(value1*VC21.Coef-value2*VC11.Coef) < 0.0000001 {
			return nil, errors.New("Infinite Number of Solutions")
		} else {
			return nil, errors.New("No Solutions")
		}
	}

	// Solve the equations
	valueForVariable1 := (value1*VC22.Coef - value2*VC12.Coef) / (VC11.Coef*VC22.Coef - VC12.Coef*VC21.Coef)
	valueForVariable2 := (value2*VC11.Coef - value1*VC21.Coef) / (VC11.Coef*VC22.Coef - VC12.Coef*VC21.Coef)

	var solution []VariableAndValue
	solution = append(solution, VariableAndValue{VC11.Variable, valueForVariable1})
	solution = append(solution, VariableAndValue{VC12.Variable, valueForVariable2})
	return solution, nil
}
