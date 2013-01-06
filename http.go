package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"strconv"
)

type homePageData struct {
	Title     string
	Neighbors Neighbors
	Data      map[string]string
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/home.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Println("Could not parse templates", err)
		fmt.Fprintln(w, "Problem parsing templates", err)
		return
	}

	data := new(homePageData)
	data.Title = "BGP Status"
	data.Neighbors = neighbors

	// fmt.Fprintf(w, "%s\t%s\t%v\n", neighbor, data.State, data.Updates)
	// fmt.Printf("TMPL %s %#v\n", tmpl, tmpl)

	tmpl.Execute(w, data)

}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	for neighbor, data := range neighbors {
		fmt.Fprintf(w, "%s\t%s\t%v\n", neighbor, data.State, data.Updates)
	}
}

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/json")

	vars := mux.Vars(r)

	neighbor_ip := vars["neighbor"]
	neighbor, ok := neighbors[neighbor_ip]
	if !ok {
		w.WriteHeader(404)
		return
	}

	// fmt.Printf("VARS: %#v\n", vars)

	switch vars["method"] {
	case "asn":
		asn, err := strconv.Atoi(vars["id"])
		if err != nil {
			fmt.Fprintln(w, "Could not parse AS number")
			return
		}
		prefixes := neighbor.AsnPrefix[ASN(asn)]

		strPrefixes := make([]string, 0)

		for prefix, _ := range prefixes {
			strPrefixes = append(strPrefixes, prefix)
		}

		json, err := json.Marshal(map[string][]string{"prefixes": strPrefixes})
		// json, err := json.Marshal(prefixes)

		fmt.Fprint(w, string(json))

	case "ip":
		fmt.Fprintf(w, "/api/ip/%s not implemented\n", vars["id"])
		return
	case "prefixes":
		prefixes := neighbor.PrefixAsn
		json, err := json.Marshal(map[string]Prefixes{"prefixes": prefixes})
		if err != nil {
			fmt.Fprint(w, "error generating json", err)
		}
		fmt.Fprint(w, string(json))
	default:
		w.WriteHeader(404)
	}
	fmt.Fprint(w, "\n")
}

func httpServer() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/api/{neighbor:[0-9.:]+}/{method:asn}/{id:[0-9]+}", ApiHandler)
	r.HandleFunc("/api/{neighbor:[0-9.:]+}/{method:ip}/{id:[0-9.:]+}", ApiHandler)
	r.HandleFunc("/api/{neighbor:[0-9.:]+}/{method:prefixes}", ApiHandler)
	r.HandleFunc("/status", StatusHandler)
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
