package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Instruction struct {
	Index     int    // número de instrucción (0,1,2,…)
	Operation string // operación (LOAD_CONST, STORE_FAST, etc.)
	Operand   string // argumento u operando (0,1,2,3… o vacío)
}

type Var struct { //Esto es para guardarlo en la memoria con el tipo y valor
	Kind  Kind
	Value interface{}
}

type Kind string // puede ser "Int", "Float", etc.

type Memory map[string]interface{} // ALmacen, nombres no repetidos, un diccionario por nombre

var MEMORY Memory = make(Memory) //Inicialización de memoria

type Stack []interface{} // Pila

var stack Stack //Inicialización de pila

type instructions []Instruction

var listInstructions instructions //Lista de instrucciones

// -------------------------------Esto es un map para guardar las funciones-----------------------------------------------
var GLOBALS = map[string]func(args ...interface{}) (interface{}, error){ //Mapa con las funciones, en este caso solo print
	"print": func(args ...interface{}) (interface{}, error) {
		fmt.Println(args...)
		return nil, nil
	},
}

// INICIO DE FUNCIONES
// -------------------------El push agrega el elemento any al final----------------------------
func (s *Stack) Push(x any) { *s = append(*s, x) }

// -------------------------El pop quita y devuelve el elemento any del final----------------------------
func (s *Stack) Pop() (interface{}, error) {
	if len(*s) == 0 {
		return nil, fmt.Errorf("stack underflow")
	}
	last := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return last, nil
}

// -----------------------------Funciones necesarias para el COMPARE_OP--------------------------------------

// Para saber si es Int
func isIntKind(k reflect.Kind) bool {
	return k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64
}

// Para saber si es Float
func isFloatKind(k reflect.Kind) bool {
	return k == reflect.Float32 || k == reflect.Float64
}
func isBoolKind(k reflect.Kind) bool {
	return k == reflect.Bool
}

// Para convertir de int to float
func toFloat64(v reflect.Value) float64 {
	if isIntKind(v.Kind()) {
		return float64(v.Int())
	}
	return v.Float()
}

// Función para aplicar cálculos numéricos
func applyGenericCalc(op string, a, b interface{}) (interface{}, error) {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if isIntKind(va.Kind()) && isIntKind(vb.Kind()) {
		af := float64(va.Int())
		bf := float64(vb.Int())
		return applyNumericCalc(op, af, bf)
	}
	if isIntKind(va.Kind()) && isFloatKind(vb.Kind()) || isIntKind(vb.Kind()) && isFloatKind(va.Kind()) {
		af := toFloat64(va)
		bf := toFloat64(vb)
		return applyNumericCalc(op, af, bf)
	}
	if isFloatKind(va.Kind()) && isFloatKind(vb.Kind()) {
		af := va.Float()
		bf := vb.Float()
		return applyNumericCalc(op, af, bf)
	}
	// Si ambos son strings → comparación o concatenación
	if va.Kind() == reflect.String && vb.Kind() == reflect.String {
		as := va.String()
		bs := vb.String()
		return applyStringCalc(op, as, bs)
	}

	return nil, fmt.Errorf("tipos incompatibles: %v y %v", va.Kind(), vb.Kind())

}

func applyStringCalc(op string, a, b string) (any, error) {
	switch op {
	case "+":
		return a + b, nil
	}
	return "", fmt.Errorf("Operación no soportada para strings %q", op)
}

func applyNumericCalc(op string, a, b float64) (any, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "/":
		if b != 0 {
			return a / b, nil
		} else {
			return fmt.Errorf("Error, no se pudo dividir %d entre %d", a, b), nil
		}

	case "*":
		return a * b, nil
	case "%":
		if b == 0 {
			return nil, fmt.Errorf("división por cero en módulo")
		}
		return int64(a) % int64(b), nil
	default:
		return nil, fmt.Errorf("operador no soportado para números: %q", op)
	}
}

// Se compara
func applyGenericOp(op string, a, b interface{}) (interface{}, error) {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// Si ambos son enteros (int, int64, etc.)
	if isIntKind(va.Kind()) && isIntKind(vb.Kind()) {
		af := float64(va.Int())
		bf := float64(vb.Int())
		return applyNumericOp(op, af, bf)
	}

	// Si uno es int y otro float → convertir ambos a float64
	if (isIntKind(va.Kind()) && isFloatKind(vb.Kind())) || (isFloatKind(va.Kind()) && isIntKind(vb.Kind())) {
		af := toFloat64(va)
		bf := toFloat64(vb)
		return applyNumericOp(op, af, bf)
	}

	// Si ambos son floats
	if isFloatKind(va.Kind()) && isFloatKind(vb.Kind()) {
		af := va.Float()
		bf := vb.Float()
		return applyNumericOp(op, af, bf)
	}

	// Si ambos son strings → comparación o concatenación
	if va.Kind() == reflect.String && vb.Kind() == reflect.String {
		as := va.String()
		bs := vb.String()
		return applyStringOp(op, as, bs)
	}

	return nil, fmt.Errorf("tipos incompatibles: %v y %v", va.Kind(), vb.Kind())
}

func applyNumericOp(op string, a, b float64) (interface{}, error) {
	switch op {
	case "+":
		return a + b, nil
	case "==":
		return a == b, nil
	case "!=":
		return a != b, nil
	case "<":
		return a < b, nil
	case ">":
		return a > b, nil
	case "<=":
		return a <= b, nil
	case ">=":
		return a >= b, nil
	default:
		return nil, fmt.Errorf("operador no soportado para números: %q", op)
	}
}

func applyStringOp(op string, a, b string) (interface{}, error) {
	switch op {
	case "+":
		return a + b, nil
	case "==":
		return a == b, nil
	case "!=":
		return a != b, nil
	case "<":
		return a < b, nil
	case ">":
		return a > b, nil
	case "<=":
		return a <= b, nil
	case ">=":
		return a >= b, nil
	default:
		return nil, fmt.Errorf("operador no soportado para strings: %q", op)
	}
}

// ----------------- LECTOR DE ARCHIVO-----------------

// readInstructions lee un archivo de texto y devuelve el slice de instrucciones.
// Formato esperado por línea: "<index> <OP> [<operand (opcional)>]"
// - líneas vacías o que comienzan con '#' se ignoran
// - usa SplitN para no reventar si el operando trae espacios (ej: strings entre comillas)

var ws = regexp.MustCompile(`\s+`) //Reduce los espacios o tabs a uno para poder usar el split

// readInstructions lee "<index> <OP> [<operand>]" y soporta múltiples espacios/tabs.
// Ignora líneas vacías o que empiezan con '#'. También corta comentarios inline.
func (listIntructions *instructions) readInstructions(path string) error { // Declara la función; recibe la ruta del archivo y devuelve instrucciones o error
	file, err := os.Open(path) // abre el path del archivo txt
	if err != nil {            // si hubo error al abrir (no existe, permisos, etc.)
		return err // retorna nil y el error
	}
	defer file.Close() // cierra el archivo cuando termine la función (éxito o fallo)

	scanner := bufio.NewScanner(file) // Scanner para leer el archivo línea por línea
	lineNo := 0                       // contador de líneas, útil para mensajes de error

	for scanner.Scan() { // recorre cada línea del archivo
		lineNo++                                  // incrementa número de línea actual
		line := strings.TrimSpace(scanner.Text()) // obtiene el texto de la línea y elimina espacios al inicio/fin

		if line == "" || strings.HasPrefix(line, "#") { // si la línea está vacía o comienza con '#', se ignora
			continue // pasa a la siguiente línea
		}
		// Comentario inline opcional: todo después de '#' se ignora
		if i := strings.Index(line, "#"); i >= 0 { // busca la primera aparición de '#'
			line = strings.TrimSpace(line[:i]) // recorta el contenido antes del '#' y trimea espacios
			if line == "" {                    // si al recortar quedó vacía,
				continue // se ignora la línea completa
			}
		}

		parts := ws.Split(line, 3) // índice, operación, (resto como operando); ws es un regexp `\s+` para separar por 1+ espacios/tabs
		if len(parts) < 2 {        // se requieren al menos 2 partes: índice y operación
			return fmt.Errorf("línea %d inválida: %q", lineNo, line) // error con número de línea y contenido
		}

		index, err := strconv.Atoi(parts[0]) // convierte la primera parte (índice) a entero
		if err != nil {                      // si no es un número válido
			return fmt.Errorf("índice inválido en línea %d: %v", lineNo, err) // retorna error detallado
		}
		op := parts[1]       // segunda parte: nombre de la operación (opcode)
		operand := ""        // inicializa el operando como cadena vacía (puede no existir)
		if len(parts) == 3 { // si hay tercera parte,
			operand = parts[2] // se preservan los espacios internos del operando tal cual
		}

		*listIntructions = append(*listIntructions, Instruction{index, op, operand}) // agrega la instrucción parseada al slice
	}

	if err := scanner.Err(); err != nil { // verifica si el Scanner encontró algún error de lectura
		return err // retorna el error del escaneo
	}
	return nil
}

// --------------------------Esta función identifica el tipo de un valor de Go y lo devuelve junto al mismo valor-----------------------
func identify_type(val interface{}) (interface{}, string) {
	var typeName string
	switch val.(type) {
	case int:
		typeName = "Int"
	case float64:
		typeName = "Float"
	case string:
		typeName = "String"
	case bool:
		typeName = "Bool"
	case nil:
		typeName = "None"
	case []interface{}:
		typeName = "List"
	default:
		typeName = "Unknown"
	}
	return val, typeName
}

// ---------------------------------- parseLiteral convierte un string textual (como "123", "True", '"Hola"') a su valor correspondiente en Go
func convert_to_value(operand string) (interface{}, error) {
	operand = strings.TrimSpace(operand) // Elimina espacios al inicio y al final

	if operand == "" {
		return nil, fmt.Errorf("literal vacío") // Si está vacío, error
	}

	// -------- STRING entre comillas simples o dobles --------
	if (strings.HasPrefix(operand, "\"") && strings.HasSuffix(operand, "\"")) ||
		(strings.HasPrefix(operand, "'") && strings.HasSuffix(operand, "'")) {

		// Intenta quitar las comillas y procesar secuencias tipo \" o \n
		unq, err := strconv.Unquote(operand)
		if err != nil {
			// Si falló por ser comillas simples, conviértelas a dobles y reintenta
			if strings.HasPrefix(operand, "'") && strings.HasSuffix(operand, "'") {
				return strconv.Unquote("\"" + operand[1:len(operand)-1] + "\"")
			}
			return nil, err
		}
		return unq, nil // Devuelve el string limpio
	}

	// -------- BOOLEANOS estilo Python --------
	if operand == "True" {
		return true, nil
	}
	if operand == "False" {
		return false, nil
	}
	// -------- NONE estilo Python (null en Go) --------
	if operand == "None" {
		return nil, nil
	}

	// -------- INT --------
	if n, err := strconv.Atoi(operand); err == nil {
		return n, nil
	}

	// -------- FLOAT --------
	if f, err := strconv.ParseFloat(operand, 64); err == nil {
		return f, nil
	}

	// -------- NO SE PUDO CONVERTIR --------
	return nil, fmt.Errorf("literal no reconocido: %q", operand)
}

//-----------------------------Funcion para recorrer las instrucciones e implementarlas-----------------------------------------------------

func (listInstructions *instructions) executeProgram() error {
	//Recorre cada instrucción linea por linea
	for index, _ := range *listInstructions {
		idx := (*listInstructions)[index].Index           //obtiene el indice de fila
		operation := (*listInstructions)[index].Operation //se obtiene la operación
		operand := (*listInstructions)[index].Operand     //se obtiene el operando
		switch operation {                                //según la operación hace los cálculos correspondientes
		case "LOAD_CONST":
			raw := strings.TrimSpace(operand)
			val, err := convert_to_value(raw)
			if err != nil {
				return fmt.Errorf("LOAD_CONST %d: %v", idx, err)
			}
			stack.Push(val) //Guarda en la pila el valor de la constante
		case "STORE_FAST":
			name := strings.TrimSpace(operand)
			v, err := stack.Pop() //Obtiene el ultimo valor de la pila
			if err != nil {
				return fmt.Errorf("STORE_FAST %d: %v", idx, err)
			}
			val, typ := identify_type(v)
			MEMORY[name] = Var{Kind: Kind(typ), Value: val} //y lo guarda en la constante

		case "LOAD_FAST":
			name := strings.TrimSpace(operand)
			var_, ok := MEMORY[name] //Obtiene el valor de la constante en memoria
			if !ok {
				return fmt.Errorf("LOAD_FAST %d: variable no encontrada: %s", idx, name)
			}
			val := var_.(Var).Value
			stack.Push(val) //Y lo guarda en la pila

		case "LOAD_GLOBAL":
			name := strings.TrimSpace(operand)
			function, ok := GLOBALS[name] //Obtiene la dirección de la función en el mapeo de funciones
			if !ok {
				return fmt.Errorf("LOAD_GLOBAL %d: función global no encontrada: %s", idx, name)
			}
			stack.Push(function) // no se ejecuta aún, solo se pone en la pila

		case "CALL_FUNCTION":
			nArgs, err := strconv.Atoi(strings.TrimSpace(operand)) // cantidad de argumentos
			if err != nil {
				return fmt.Errorf("CALL_FUNCTION %d: argumento inválido: %v", idx, err)
			}
			// Extraer los argumentos desde la pila, en orden inverso
			args := make([]interface{}, nArgs)
			for i := nArgs - 1; i >= 0; i-- {
				arg, err := stack.Pop()
				if err != nil {
					return fmt.Errorf("CALL_FUNCTION %d: faltan argumentos: %v", idx, err)
				}
				args[i] = arg // los meto en orden correcto
			}
			// Extraer la función
			fnRaw, err := stack.Pop()
			if err != nil {
				return fmt.Errorf("CALL_FUNCTION %d: no hay función en la pila: %v", idx, err)
			}
			//Verificar si de verdad la función es de tipo función
			fn, ok := fnRaw.(func(...interface{}) (interface{}, error))
			if !ok {
				return fmt.Errorf("CALL_FUNCTION %d: tipo inválido de función: %T", idx, fnRaw)
			}
			// Llamar a la función
			result, err := fn(args...)
			if err != nil {
				return fmt.Errorf("CALL_FUNCTION %d: error en la función: %v", idx, err)

			}
			if result != nil {
				// Guardar el resultado en la pila si no es nil
				stack.Push(result)
				//fmt.Println(result)
			}

		case "COMPARE_OP":
			name := strings.TrimSpace(operand)
			oper1, err := stack.Pop() //Saca el primer operando
			if err != nil {
				return fmt.Errorf("COMPARE_OP %d: %v", idx, err)
			}
			oper2, err2 := stack.Pop() //Saca el segundo operando
			if err != nil {
				return fmt.Errorf("COMPARE_OP %d: %v", idx, err2)
			}
			switch name {
			//Según el operando de la instrucción realiza la operación
			case "<":
				okLess, err := applyGenericOp("<", oper2, oper1)
				if err != nil {
					return fmt.Errorf("OPER '<': %v", err)
				}
				stack.Push(okLess) //Guarda el valor true or false en pila
			case ">":
				okMore, err := applyGenericOp(">", oper2, oper1)
				if err != nil {
					return fmt.Errorf("OPER '>': %v", err)
				}
				stack.Push(okMore) //Guarda el valor true or false en pila

			case "==":
				okEqual, err := applyGenericOp("==", oper2, oper1)
				if err != nil {
					return fmt.Errorf("OPER '==': %v", err)
				}
				stack.Push(okEqual) //Guarda el valor true or false en pila
			case "!=":
				okNEqual, err := applyGenericOp("!=", oper2, oper1)
				if err != nil {
					return fmt.Errorf("OPER '!=': %v", err)
				}
				stack.Push(okNEqual) //Guarda el valor true or false en pila
			case "<=":
				okEless, err := applyGenericOp("<=", oper2, oper1)
				if err != nil {
					return fmt.Errorf("OPER '<=': %v", err)
				}
				stack.Push(okEless) //Guarda el valor true or false en pila
			case ">=":
				okEmore, err := applyGenericOp(">=", oper2, oper1)
				if err != nil {
					return fmt.Errorf("OPER '>=': %v", err)
				}
				stack.Push(okEmore) //Guarda el valor true or false en pila
			default:
				return fmt.Errorf("COMPARE_OP %d: operador no soportado: %q", idx, name)
			}
		case "BINARY_SUBSTRACT":
			oper1, err := stack.Pop() //Saca el primer operando
			if err != nil {
				return fmt.Errorf("BINARY_SUBSTRACT %d: %v", idx, err)
			}
			oper2, err := stack.Pop() //Saca el segundo operando
			if err != nil {
				return fmt.Errorf("BINARY_SUBSTRACT %d: %v", idx, err)
			}
			okSum, err := applyGenericCalc("-", oper2, oper1)

			if err != nil {
				return fmt.Errorf("BINARY_SUBSTRACT %d: %v", idx, err)
			}
			stack.Push(okSum)
		case "BINARY_ADD":
			oper1, err := stack.Pop() //Saca el primer operando
			if err != nil {
				return fmt.Errorf("BINARY_ADD %d: %v", idx, err)
			}
			oper2, err := stack.Pop() //Saca el segundo operando
			if err != nil {
				return fmt.Errorf("BINARY_ADD %d: %v", idx, err)
			}
			okSum, err := applyGenericCalc("+", oper2, oper1)

			if err != nil {
				return fmt.Errorf("BINARY_ADD %d: %v", idx, err)
			}
			stack.Push(okSum)

		case "BINARY_DIVIDE":
			oper1, err := stack.Pop() //Saca el primer operando
			if err != nil {
				return fmt.Errorf("BINARY_DIVIDE\n %d: %v", idx, err)
			}
			oper2, err := stack.Pop() //Saca el segundo operando
			if err != nil {
				return fmt.Errorf("BINARY_DIVIDE\n %d: %v", idx, err)
			}
			okDiv, err := applyGenericCalc("/", oper2, oper1)
			if err != nil {
				return fmt.Errorf("BINARY_DIVIDE\n %d: %v", idx, err)
			}
			stack.Push(okDiv)
		case "BINARY_MULTIPLY":
			oper1, err := stack.Pop() //Saca el primer operando
			if err != nil {
				return fmt.Errorf("BINARY_MULTIPLY\n %d: %v", idx, err)
			}
			oper2, err := stack.Pop() //Saca el segundo operando
			if err != nil {
				return fmt.Errorf("BINARY_MULTIPLY\n %d: %v", idx, err)
			}
			okSum, err := applyGenericCalc("*", oper2, oper1)
			if err != nil {
				return fmt.Errorf("BINARY_MULTIPLY\n %d: %v", idx, err)
			}
			stack.Push(okSum)
		case "BINARY_AND":
			// Pop right, then left (top of stack is right operand)
			right, err := stack.Pop()
			if err != nil {
				return fmt.Errorf("BINARY_AND %d: %v", idx, err)
			}
			left, err := stack.Pop()
			if err != nil {
				return fmt.Errorf("BINARY_AND %d: %v", idx, err)
			}

			va := reflect.ValueOf(left)
			vb := reflect.ValueOf(right)

			// bool & bool → bool (lógico AND)
			if isBoolKind(va.Kind()) && isBoolKind(vb.Kind()) {
				stack.Push(va.Bool() && vb.Bool())
				break
			}

			// Admite int-like y bool mezclados (bool como 0/1)
			aOK := isIntKind(va.Kind()) || isBoolKind(va.Kind())
			bOK := isIntKind(vb.Kind()) || isBoolKind(vb.Kind())
			if aOK && bOK {
				// Normalizamos a int64 para operar bit a bit
				var ai, bi int64
				if isIntKind(va.Kind()) {
					ai = va.Int()
				} else { // bool
					if va.Bool() {
						ai = 1
					} else {
						ai = 0
					}
				}
				if isIntKind(vb.Kind()) {
					bi = vb.Int()
				} else { // bool
					if vb.Bool() {
						bi = 1
					} else {
						bi = 0
					}
				}

				res := ai & bi

				// Si ambos operandos eran bool, ya habríamos retornado arriba.
				// Aquí devolvemos int (estilo Python cuando hay ints de por medio).
				stack.Push(int(res))
				break
			}

			return fmt.Errorf("BINARY_AND %d: tipos no soportados %T & %T", idx, left, right)

		case "BINARY_OR":
			right, err := stack.Pop()
			if err != nil {
				return fmt.Errorf("BINARY_OR %d: %v", idx, err)
			}
			left, err := stack.Pop()
			if err != nil {
				return fmt.Errorf("BINARY_OR %d: %v", idx, err)
			}

			va := reflect.ValueOf(left)
			vb := reflect.ValueOf(right)

			// bool | bool → bool (lógico OR)
			if isBoolKind(va.Kind()) && isBoolKind(vb.Kind()) {
				stack.Push(va.Bool() || vb.Bool())
				break
			}

			// Admite int-like y bool mezclados (bool como 0/1)
			aOK := isIntKind(va.Kind()) || isBoolKind(va.Kind())
			bOK := isIntKind(vb.Kind()) || isBoolKind(vb.Kind())
			if aOK && bOK {
				var ai, bi int64
				if isIntKind(va.Kind()) {
					ai = va.Int()
				} else {
					if va.Bool() {
						ai = 1
					} else {
						ai = 0
					}
				}
				if isIntKind(vb.Kind()) {
					bi = vb.Int()
				} else {
					if vb.Bool() {
						bi = 1
					} else {
						bi = 0
					}
				}

				res := ai | bi
				stack.Push(int(res))
				break
			}

			return fmt.Errorf("BINARY_OR %d: tipos no soportados %T | %T", idx, left, right)

		case "BINARY_MODULO":
			oper1, err := stack.Pop() //Saca el primer operando
			if err != nil {
				return fmt.Errorf("BINARY_MODULO\n %d: %v", idx, err)
			}
			oper2, err := stack.Pop() //Saca el segundo operando
			if err != nil {
				return fmt.Errorf("BINARY_MODULO\n %d: %v", idx, err)
			}
			okMod, err := applyGenericCalc("%", oper2, oper1)
			if err != nil {
				return fmt.Errorf("BINARY_MODULO\n %d: %v", idx, err)
			}
			stack.Push(okMod)
		case "STORE_SUBSCR":
			// Espera en la pila: [index, array, value] (value en el tope)
			value, err := stack.Pop()
			if err != nil {
				return fmt.Errorf("STORE_SUBSCR %d: %v", idx, err)
			}
			arrRaw, err := stack.Pop()
			if err != nil {
				return fmt.Errorf("STORE_SUBSCR %d: %v", idx, err)
			}
			idxRaw, err := stack.Pop()
			if err != nil {
				return fmt.Errorf("STORE_SUBSCR %d: %v", idx, err)
			}
			arr, ok := arrRaw.([]interface{})
			if !ok {
				return fmt.Errorf("STORE_SUBSCR %d: tipo de arreglo inválido: %T", idx, arrRaw)
			}
			// Convertir índice a int
			var i int
			switch v := idxRaw.(type) {
			case int:
				i = v
			case float64:
				i = int(v)
			default:
				return fmt.Errorf("STORE_SUBSCR %d: índice inválido de tipo %T", idx, idxRaw)
			}
			if i < 0 || i >= len(arr) {
				return fmt.Errorf("STORE_SUBSCR %d: índice fuera de rango: %d", idx, i)
			}
			arr[i] = value

		case "BINARY_SUBSCR":
			// Espera en la pila: [index, array] (array en el tope)
			arrRaw, err := stack.Pop()
			if err != nil {
				return fmt.Errorf("BINARY_SUBSCR %d: %v", idx, err)
			}
			idxRaw, err := stack.Pop()
			if err != nil {
				return fmt.Errorf("BINARY_SUBSCR %d: %v", idx, err)
			}
			arr, ok := arrRaw.([]interface{})
			if !ok {
				return fmt.Errorf("BINARY_SUBSCR %d: tipo de arreglo inválido: %T", idx, arrRaw)
			}
			var i int
			switch v := idxRaw.(type) {
			case int:
				i = v
			case float64:
				i = int(v)
			default:
				return fmt.Errorf("BINARY_SUBSCR %d: índice inválido de tipo %T", idx, idxRaw)
			}
			if i < 0 || i >= len(arr) {
				return fmt.Errorf("BINARY_SUBSCR %d: índice fuera de rango: %d", idx, i)
			}
			stack.Push(arr[i])

		case "JUMP_ABSOLUTE":
			// Nota: Para realizar el salto, ejecutamos recursivamente desde 'target'.
			targetStr := strings.TrimSpace(operand)
			target, err := strconv.Atoi(targetStr)
			if err != nil {
				return fmt.Errorf("JUMP_ABSOLUTE %d: target inválido: %v", idx, err)
			}
			// Buscar la posición 'pos' en el slice donde Index == target
			pos := -1
			for p := range *listInstructions {
				if (*listInstructions)[p].Index == target {
					pos = p
					break
				}
			}
			if pos == -1 {
				return fmt.Errorf("JUMP_ABSOLUTE %d: target %d no encontrado", idx, target)
			}
			// Crear un subprograma desde 'pos' y ejecutarlo
			sub := make(instructions, 0, len(*listInstructions))
			sub = append(sub, (*listInstructions)[pos:]...)
			sub = append(sub, (*listInstructions)[:pos]...)
			return (&sub).executeProgram()

		case "JUMP_IF_TRUE":
			// Si el valor del tope es true, saltar a 'target'
			val, err := stack.Pop()
			if err != nil {
				return fmt.Errorf("JUMP_IF_TRUE %d: %v", idx, err)
			}
			b, ok := val.(bool)
			if !ok {
				return fmt.Errorf("JUMP_IF_TRUE %d: valor no booleano: %T", idx, val)
			}
			targetStr := strings.TrimSpace(operand)
			target, err := strconv.Atoi(targetStr)
			if err != nil {
				return fmt.Errorf("JUMP_IF_TRUE %d: target inválido: %v", idx, err)
			}
			if b {
				pos := -1
				for p := range *listInstructions {
					if (*listInstructions)[p].Index == target {
						pos = p
						break
					}
				}
				if pos == -1 {
					return fmt.Errorf("JUMP_IF_TRUE %d: target %d no encontrado", idx, target)
				}
				// Rotar el programa: [pos:...] + [:pos]  (incluye TODO, no recorta)
				sub := make(instructions, 0, len(*listInstructions))
				sub = append(sub, (*listInstructions)[pos:]...)
				sub = append(sub, (*listInstructions)[:pos]...)
				return (&sub).executeProgram()

			}

		case "JUMP_IF_FALSE":
			// Si el valor del tope es false, saltar a 'target'
			val, err := stack.Pop()
			if err != nil {
				return fmt.Errorf("JUMP_IF_FALSE %d: %v", idx, err)
			}
			b, ok := val.(bool)
			if !ok {
				return fmt.Errorf("JUMP_IF_FALSE %d: valor no booleano: %T", idx, val)
			}
			targetStr := strings.TrimSpace(operand)
			target, err := strconv.Atoi(targetStr)
			if err != nil {
				return fmt.Errorf("JUMP_IF_FALSE %d: target inválido: %v", idx, err)
			}
			if !b {
				pos := -1
				for p := range *listInstructions {
					if (*listInstructions)[p].Index == target {
						pos = p
						break
					}
				}
				if pos == -1 {
					return fmt.Errorf("JUMP_IF_FALSE %d: target %d no encontrado", idx, target)
				}
				// Rotar el programa: [pos:...] + [:pos]  (incluye TODO, no recorta)
				sub := make(instructions, 0, len(*listInstructions))
				sub = append(sub, (*listInstructions)[pos:]...)
				sub = append(sub, (*listInstructions)[:pos]...)
				return (&sub).executeProgram()

			}

		case "BUILD_LIST":
			// Construye una lista con 'elements' elementos desde la pila
			n, err := strconv.Atoi(strings.TrimSpace(operand))
			if err != nil {
				return fmt.Errorf("BUILD_LIST %d: parámetro inválido: %v", idx, err)
			}
			if n < 0 {
				return fmt.Errorf("BUILD_LIST %d: cantidad negativa: %d", idx, n)
			}
			list := make([]interface{}, n)
			for i := n - 1; i >= 0; i-- {
				v, err := stack.Pop()
				if err != nil {
					return fmt.Errorf("BUILD_LIST %d: faltan elementos: %v", idx, err)
				}
				list[i] = v
			}
			stack.Push(list)

		case "END":
			// Termina la ejecución del programa
			return nil

		default:
			return fmt.Errorf("Error en la linea %d:  funcion no existente", idx)
		}
		fmt.Println(stack)
		fmt.Println(MEMORY)
	}
	return nil
}

func main() {
	//Se cargan las instrucciones en la lista listInstructions
	//Acá se cambia el tipo de archivo dependiendo de la prueba que se quiera usar
	err := listInstructions.readInstructions("08_loop_counter.txt")
	if err != nil {
		fmt.Errorf("", err)
	}

	// Se muestran las instrucciones cargadas"
	for _, ins := range listInstructions {
		fmt.Printf("%d\t%s\t%s\n", ins.Index, ins.Operation, ins.Operand)

	}
	//Imprime el error si hay alguno con el código
	if err := listInstructions.executeProgram(); err != nil {
		fmt.Println("error:", err)
	}

}
