package functions

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"gone-be/src/utils"
)

// TestGraphQL is a test function for GraphQL
func TestGraphQL() (map[string]interface{}, *utils.ServiceError) {
	// Define schema fields
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
	}

	// Create schema
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return nil, &utils.ServiceError{
			StatusCode: 500,
			Message:    fmt.Sprintf("failed to create schema: %v", err),
		}
	}

	// Define query
	query := `
		{
			hello
		}
	`

	// Execute query
	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		return nil, &utils.ServiceError{
			StatusCode: 400,
			Message:    fmt.Sprintf("GraphQL execution error: %+v", r.Errors),
		}
	}

	// Ensure r.Data is of correct type
	response, ok := r.Data.(map[string]interface{})
	if !ok {
		return nil, &utils.ServiceError{
			StatusCode: 500,
			Message:    "unexpected response format: data is not a map",
		}
	}

	return response, nil
}
