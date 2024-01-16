package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/graphql-go/graphql"
)

// Definición de la estructura Persona
type Persona struct {
	ID     string  `json:"id"`
	Nombre *string `json:"nombre"`
	Edad   int     `json:"edad"`
	Ciudad *string `json:"ciudad"`
	Genero *string `json:"genero"`
}

// Datos ficticios para personas
var personas = []Persona{
	{"1", strPtr("Juan"), 25, strPtr("Ciudad A"), strPtr("Masculino")},
	{"2", strPtr("María"), 30, strPtr("Ciudad B"), strPtr("Femenino")},
	{"6", strPtr("María"), 30, strPtr("Ciudad B"), strPtr("Femenino")},
	{"3", strPtr("Carlos"), 22, strPtr("Ciudad C"), strPtr("Masculino")},
	{"4", strPtr("Laura"), 28, strPtr("Ciudad A"), strPtr("Femenino")},
	{"5", strPtr("Pedro"), 35, strPtr("Ciudad B"), strPtr("Masculino")},
}

func strPtr(s string) *string {
	return &s
}

// Definición del tipo GraphQL para Persona
var personaType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Persona",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"nombre": &graphql.Field{
				Type: graphql.String,
			},
			"edad": &graphql.Field{
				Type: graphql.Int,
			},
			"ciudad": &graphql.Field{
				Type: graphql.String,
			},
			"genero": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

// Definición del esquema GraphQL
var schema, err = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: graphql.NewObject(
			graphql.ObjectConfig{
				Name: "Query",
				Fields: graphql.Fields{
					// Nuevo campo para obtener información de rango de edades por ciudad
					"infoFiltrada": &graphql.Field{
						Type: graphql.NewList(personaType),
						Args: graphql.FieldConfigArgument{
							"edadMin": &graphql.ArgumentConfig{
								Type: graphql.Int,
							},
							"edadMax": &graphql.ArgumentConfig{
								Type: graphql.Int,
							},
							"ciudad": &graphql.ArgumentConfig{
								Type: graphql.String,
							},
							"genero": &graphql.ArgumentConfig{
								Type: graphql.String,
							},
						},
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							// Filtrar personas según los parámetros proporcionados
							var personasFiltradas []Persona
							for _, persona := range personas {
								edadValida := true
								ciudadValida := true
								generoValido := true

								// Filtrar por edad si se proporciona
								if edadMin, ok := p.Args["edadMin"]; ok && edadMin != nil {
									if edad, castOk := edadMin.(int); castOk && persona.Edad < edad {
										edadValida = false
									}
								}

								if edadMax, ok := p.Args["edadMax"]; ok && edadMax != nil {
									if edad, castOk := edadMax.(int); castOk && persona.Edad > edad {
										edadValida = false
									}
								}

								// Filtrar por ciudad si se proporciona
								if ciudad, ok := p.Args["ciudad"]; ok {
									if ciudad != nil {
										if ciudadStr, castOk := ciudad.(string); castOk && persona.Ciudad != nil && *persona.Ciudad != ciudadStr {
											ciudadValida = false
										}
									} else {
										// Si se proporciona ciudad nula, filtrar personas con ciudad no nula
										if persona.Ciudad != nil {
											ciudadValida = false
										}
									}
								}

								// Filtrar por género si se proporciona
								if genero, ok := p.Args["genero"]; ok && genero != nil {
									if generoStr, castOk := genero.(*string); castOk && persona.Genero != generoStr {
										generoValido = false
									}
								}

								// Agregar persona a la lista si pasa todos los filtros
								if edadValida && ciudadValida && generoValido {
									personasFiltradas = append(personasFiltradas, persona)
								}
							}

							return personasFiltradas, nil

						},
					},
				},
			},
		),
	},
)

func graphqlHandler(w http.ResponseWriter, r *http.Request) {
	// Permitir solicitudes CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Comprobar el método de solicitud para las solicitudes CORS preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Comprobar si la solicitud es GET
	if r.Method == "GET" {
		// Obtener la consulta de la URL
		queryParam := r.URL.Query().Get("query")

		// Manejar las consultas GraphQL
		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: queryParam,
		})

		// Formatear la salida JSON con sangrías y líneas nuevas
		formattedResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			http.Error(w, "Error al formatear la respuesta JSON", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(formattedResult)
		return
	}

	// Comprobar si la solicitud es POST
	if r.Method == "POST" {
		// Leer el cuerpo de la solicitud
		var requestData map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Error al leer el cuerpo de la solicitud", http.StatusBadRequest)
			return
		}

		// Obtener la consulta del cuerpo de la solicitud
		query, ok := requestData["query"].(string)
		if !ok {
			http.Error(w, "Consulta no proporcionada en el cuerpo de la solicitud", http.StatusBadRequest)
			return
		}

		// Remover valores nulos de la consulta GraphQL antes de ejecutarla
		queryWithoutNulls := removeNullsFromQuery(query, requestData)

		// Manejar las consultas GraphQL
		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: queryWithoutNulls,
		})

		// Formatear la salida JSON con sangrías y líneas nuevas
		formattedResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			http.Error(w, "Error al formatear la respuesta JSON", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(formattedResult)
		return
	}

	// Método de solicitud no admitido
	http.Error(w, "Método de solicitud no admitido", http.StatusMethodNotAllowed)
}

// Función para quitar valores nulos de la consulta GraphQL
func removeNullsFromQuery(query string, requestData map[string]interface{}) string {
	for key, value := range requestData {
		if value == nil {
			// Remover el argumento completo, incluyendo el valor nulo
			query = removeArgument(query, key)
		}
	}
	return query
}

// Función para remover un argumento completo de la consulta GraphQL
func removeArgument(query string, argName string) string {
	// Buscar el argumento en la consulta
	argStart := strings.Index(query, fmt.Sprintf("%s:", argName))
	if argStart == -1 {
		return query
	}

	// Encontrar el final del argumento
	bracketCount := 0
	argEnd := argStart
	for argEnd < len(query) {
		if query[argEnd] == '(' {
			bracketCount++
		} else if query[argEnd] == ')' {
			bracketCount--
			if bracketCount == 0 {
				break
			}
		}
		argEnd++
	}

	// Remover el argumento completo de la consulta
	query = query[:argStart] + query[argEnd+1:]

	return query
}

func main() {
	// Configurar el manejador GraphQL
	http.HandleFunc("/graphql", graphqlHandler)

	// Iniciar el servidor en el puerto 8080
	fmt.Println("Servidor GraphQL en http://localhost:8080/graphql")
	http.ListenAndServe(":8080", nil)
}
