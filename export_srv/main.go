package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/mei-rune/httpdump"
)

func main() {
	var dir string
	var listen string

	flag.StringVar(&dir, "dir", "", "")
	flag.StringVar(&listen, "listen", ":", "")
	flag.Parse()

	srv, err := httpdump.NewServer(dir)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = http.ListenAndServe(listen, http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		// switch request.URL.Path {
		// case "/api/v32/timeseries":
		// 	queryParams := request.URL.Query()

		// 	q, err := ParseQuery(queryParams.Get("query"))
		// 	if err != nil {
		// 		http.Error(response, err.Error(), http.StatusInternalServerError)
		// 	} else {
		// 		query
		// 	}
		// 	return 
		// }
		srv.ServeHTTP(response, request)
	}))
	if err != nil {
		fmt.Println(err)
		return
	}
}
