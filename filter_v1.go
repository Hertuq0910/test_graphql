// FILTRO POR RANGO DE EDADES
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"
)

// Definición de la estructura Persona
type Persona struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	Edad   int    `json:"edad"`
	Ciudad string `json:"ciudad"`
}

// Datos ficticios para personas
var personas = []Persona{
	{"1", "Juan", 25, "Ciudad A"},
	{"2", "María", 30, "Ciudad B"},
	{"3", "Carlos", 22, "Ciudad C"},
	{"4", "Laura", 28, "Ciudad A"},
	{"5", "Pedro", 35, "Ciudad B"},
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
		},
	},
)

// Definición del esquema GraphQL
var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: graphql.NewObject(
			graphql.ObjectConfig{
				Name: "Query",
				Fields: graphql.Fields{
					"persona": &graphql.Field{
						Type: personaType,
						Args: graphql.FieldConfigArgument{
							"id": &graphql.ArgumentConfig{
								Type: graphql.String,
							},
						},
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							id, ok := p.Args["id"].(string)
							if ok {
								// Buscar la persona por ID
								for _, persona := range personas {
									if persona.ID == id {
										return persona, nil
									}
								}
							}
							return nil, nil
						},
					},
					// Agregar el nuevo campo "personasEnRangoDeEdad"
					"personasEnRangoDeEdad": &graphql.Field{
						Type: graphql.NewList(personaType),
						Args: graphql.FieldConfigArgument{
							"edadMin": &graphql.ArgumentConfig{
								Type: graphql.Int,
							},
							"edadMax": &graphql.ArgumentConfig{
								Type: graphql.Int,
							},
						},
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							edadMin, edadMinOK := p.Args["edadMin"].(int)
							edadMax, edadMaxOK := p.Args["edadMax"].(int)

							if edadMinOK && edadMaxOK {
								// Buscar personas cuyas edades estén en el rango especificado
								personasEnRango := make([]Persona, 0)
								for _, persona := range personas {
									if persona.Edad >= edadMin && persona.Edad <= edadMax {
										personasEnRango = append(personasEnRango, persona)
									}
								}
								return personasEnRango, nil
							}

							return nil, nil
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

		// Manejar las consultas GraphQL
		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: query,
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

func main() {
	// Configurar el manejador GraphQL
	http.HandleFunc("/graphql", graphqlHandler)

	// Iniciar el servidor en el puerto 8080
	fmt.Println("Servidor GraphQL en http://localhost:8080/graphql")
	http.ListenAndServe(":8080", nil)
}
