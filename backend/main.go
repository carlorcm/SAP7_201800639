package main

import (
    "fmt"
    "net/http"
	"encoding/json"
	"io/ioutil"
    "github.com/rs/cors"
	"database/sql"
    _ "github.com/go-sql-driver/mysql"
    "log"
    "os"
)

var db *sql.DB

func main() {

    

    mux := http.NewServeMux()

    mux.HandleFunc("/calculate", calculateHandler)

    mux.HandleFunc("/resultados", resultHandler)

    handler := cors.Default().Handler(mux)
    http.ListenAndServe(":8080", handler)
}

type CalculationRequest struct {
    Value1   float64 `json:"value1"`
    Value2   float64 `json:"value2"`
    Operator string  `json:"operator"`
}

// Estructura para guardar los datos de la tabla resultados
type Resultado struct {
	ID       int     `json:"id"`
	Value1   float64 `json:"value1"`
	Value2   float64 `json:"value2"`
	Operator string  `json:"operator"`
	Result   float64 `json:"result"`
	Date     string  `json:"date"`
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
    // Leer el cuerpo de la solicitud
    requestBody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error al leer el cuerpo de la solicitud", http.StatusBadRequest)
        return
    }

    // Decodificar el cuerpo de la solicitud en una instancia de CalculationRequest
    var calculationRequest CalculationRequest
    err = json.Unmarshal(requestBody, &calculationRequest)
    if err != nil {
        http.Error(w, "Error al decodificar el cuerpo de la solicitud", http.StatusBadRequest)
        return
    }

    // Realizar la operación correspondiente
    var result float64
    switch calculationRequest.Operator {
    case "+":
        result = calculationRequest.Value1 + calculationRequest.Value2
    case "-":
        result = calculationRequest.Value1 - calculationRequest.Value2
    case "*":
        result = calculationRequest.Value1 * calculationRequest.Value2
    case "/":
        result = calculationRequest.Value1 / calculationRequest.Value2
    default:
        http.Error(w, "Operador no válido", http.StatusBadRequest)
        return
    }

    // Abrir la conexión a la base de datos
    db, err := sql.Open("mysql", "root:secret@tcp(192.168.1.25:33061)/P1SO")

    if err != nil {
        panic(err)
    }

    defer db.Close()

    // Preparar la sentencia SQL INSERT
    stmt, err := db.Prepare("INSERT INTO resultados (value1, value2, operator, result) VALUES (?, ?, ?, ?)")

    if err != nil {
        panic(err.Error())
    }

    defer stmt.Close()

    // Ejecutar la sentencia SQL INSERT
    if calculationRequest.Operator == "/" && calculationRequest.Value2 == 0{
        result = -0
        _, err = stmt.Exec(calculationRequest.Value1, calculationRequest.Value2, "E", 0)

    }else{
        _, err = stmt.Exec(calculationRequest.Value1, calculationRequest.Value2, calculationRequest.Operator, result)
    }

    if err != nil {
        panic(err.Error())
    }

    // Hacer la consulta SQL
    rows, err := db.Query("SELECT * FROM resultados")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    // Crear un archivo para escribir los datos
    file, err := os.Create("/logs/resultados.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

     // Escribir los resultados en el archivo de texto
    for rows.Next() {
        var id int
        var value1 float64
        var value2 float64
        var operator string
        var result float64
        var date string
        err = rows.Scan(&id, &value1, &value2, &operator, &result, &date)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Fprintf(file, "id=%d value1=%.2f value2=%.2f operator=%s result=%.2f date=%s\n", id, value1, value2, operator, result, date)
    }
    if err := rows.Err(); err != nil {
        log.Fatal(err)
    }


    // Enviar la respuesta
    response := struct {
        Result float64 `json:"result"`
    }{result}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
    // Abrir la conexión a la base de datos
    db, err := sql.Open("mysql", "root:secret@tcp(192.168.1.25:33061)/P1SO")

    if err != nil {
        panic(err)
    }

    defer db.Close()

    // Hacer la consulta a la base de datos
    rows, err := db.Query("SELECT * FROM resultados")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    // Recorrer los resultados y guardarlos en un slice de Resultado
    var resultados []Resultado
    for rows.Next() {
        var resultado Resultado
        err := rows.Scan(&resultado.ID, &resultado.Value1, &resultado.Value2, &resultado.Operator, &resultado.Result, &resultado.Date)
        if err != nil {
            log.Fatal(err)
        }
        resultados = append(resultados, resultado)
    }
    if err := rows.Err(); err != nil {
        log.Fatal(err)
    }

    // Convertir el slice de Resultado a json
    resultadosJson, err := json.Marshal(resultados)
    if err != nil {
        log.Fatal(err)
    }

    // Escribir el json como respuesta del endpoint
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprint(w, string(resultadosJson))
}
