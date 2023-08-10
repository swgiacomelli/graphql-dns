package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	flag "github.com/spf13/pflag"

	"github.com/graphql-go/graphql"
	"github.com/sirupsen/logrus"
)

var (
	log      = logrus.New()
	logLevel = "info"
	port     = 9339
)

type hostname struct {
	IP   string `json:"id"`
	Name string `json:"name"`
}

var hostnameType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Hostname",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"ip": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"hostname": &graphql.Field{
				Type: hostnameType,
				Args: graphql.FieldConfigArgument{
					"ip": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ip, ok := p.Args["ip"].(string)
					if ok {
						return getHostname(ip)
					}
					return nil, nil
				},
			},
		},
	},
)

func getHostname(ip string) (*hostname, error) {
	if names, err := net.LookupAddr(ip); err == nil && len(names) > 0 {
		return &hostname{
			Name: names[0],
			IP:   ip,
		}, nil
	} else {
		log.Error(err)
		return nil, err
	}
}

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: queryType,
	},
)

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %v", result.Errors)
	}
	return result
}

func init() {
	flag.IntVarP(&port,
		"port",
		"p",
		9339,
		"port to export graphql")
	flag.StringVarP(&logLevel,
		"log-level", "l",
		"info",
		"log level (debug, info, warn, error, fatal, panic)")
	flag.Parse()

	if lvl, err := logrus.ParseLevel(logLevel); err == nil {
		log.SetLevel(lvl)
	} else {
		log.SetLevel(logrus.InfoLevel)
		log.Trace("Invalid log level specified (", logLevel, "), defaulting to info")
	}
}

func main() {
	log.Info("Starting graphql server on port ", port)

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		result := executeQuery(r.URL.Query().Get("query"), schema)
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Error(err)
		}
	})

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
