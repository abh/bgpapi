package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type homePageData struct {
	Title     string
	Neighbors Neighbors
	Data      map[string]string
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index").Parse(index_tpl)
	if err != nil {
		log.Println("Could not parse template", err)
		fmt.Fprintln(w, "Problem parsing template", err)
		return
	}

	neighbors_lock.RLock()
	defer neighbors_lock.RUnlock()

	data := new(homePageData)
	data.Title = "BGP Status"
	data.Neighbors = neighbors

	// fmt.Fprintf(w, "%s\t%s\t%v\n", neighbor, data.State, data.Updates)
	// fmt.Printf("TMPL %s %#v\n", tmpl, tmpl)

	tmpl.Execute(w, data)

}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	neighbors_lock.RLock()
	defer neighbors_lock.RUnlock()

	for neighbor, data := range neighbors {
		data.lock.RLock()
		defer data.lock.RUnlock()
		fmt.Fprintf(w, "%s\t%s\t%v\n", neighbor, data.State, data.Updates)
	}
}

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/json")

	vars := mux.Vars(r)

	neighbors_lock.RLock()
	defer neighbors_lock.RUnlock()

	neighbor_ip := vars["neighbor"]
	neighbor, ok := neighbors[neighbor_ip]
	if !ok {
		w.WriteHeader(404)
		return
	}

	neighbor.lock.RLock()
	defer neighbor.lock.RUnlock()

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

const index_tpl = `<!DOCTYPE html>
<html>
	<head><title>BGP Status</title>
		<link href="http://st.pimg.net/cdn/libs/bootstrap/2/css/bootstrap.min.css" rel="stylesheet">
		<style>
			html,
			body {
			  margin: 10px;
			  margin-top: 20px;
			}
		</style>
	</head>
	<body>

	<h1>{{.Title}}</h1>

	<table class="table table-striped table-hover table-condensed">
		<thead>
			<tr>
				<th>IP</th> 
				<th>State</th>
				<th>Prefixes</th>
				<th>ASNs</th>
				<th>Updates</th>
			</tr>
		</thead>

		<tbody>
		{{range $ip, $data := .Neighbors}}
			<tr>
				<td>
					{{$ip}}
				</td>
				<td>
					{{$data.State}}
				</td>
				<td>
					<a href="/api/{{$ip}}/prefixes">{{$data.PrefixCount}}</a>
				</td>
				<td>
					{{$data.AsnCount}}
				</td>
				<td>
					{{$data.Updates}}
				</td>
			<tr>
		{{else}}
			<tr><td>No neighbors</td></tr>
		{{end}}
		</tbody>
	</table>

	</body>
</html>
`
