package main

import (
	"encoding/json"
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

type postData struct {
	Query     string                 `json:"query"`
	Operation string                 `json:"operation"`
	Variables map[string]interface{} `json:"variables"`
}

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
		Name: "RootQuery",
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

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info(r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	log.Info("Starting graphql server on port ", port)

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		var p postData
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result := graphql.Do(graphql.Params{
			Context:        r.Context(),
			Schema:         schema,
			RequestString:  p.Query,
			VariableValues: p.Variables,
			OperationName:  p.Operation,
		})

		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Error(err)
		}
	})

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), logRequest(http.DefaultServeMux)))
}
