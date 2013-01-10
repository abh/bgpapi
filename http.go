package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net"
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
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)

	neighbors_lock.RLock()
	defer neighbors_lock.RUnlock()

	type ipasnResult struct {
		ASN  ASN
		Name string
	}

	switch vars["method"] {
	case "ipasn":
		_ = r.ParseForm()
		ip := net.ParseIP(r.Form.Get("ip"))

		if ip == nil {
			w.WriteHeader(400)
			fmt.Fprintln(w, "Bad IP address")
			return
		}

		result := make(map[string]*ipasnResult)

		for neighbor, data := range neighbors {
			data.lock.RLock()
			defer data.lock.RUnlock()
			result[neighbor] = new(ipasnResult)
			result[neighbor].ASN = data.FindAsn(&ip)
			result[neighbor].Name = ""
		}

		json, err := json.Marshal(map[string]interface{}{"result": result})
		if err != nil {
			w.WriteHeader(500)
			log.Println("Error generating json", err)
			fmt.Fprintln(w, "Could not generate JSON")
		}

		fmt.Fprint(w, string(json))

	default:
		w.WriteHeader(404)
		return
	}

}

func ApiNeighborHandler(w http.ResponseWriter, r *http.Request) {
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
		if err != nil {
			w.WriteHeader(500)
			log.Println("Error generating json", err)
			fmt.Fprintln(w, "Could not generate JSON")
		}
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
	r.HandleFunc("/api/{neighbor:[0-9.:]+}/{method:asn}/{id:[0-9]+}", ApiNeighborHandler)
	r.HandleFunc("/api/{neighbor:[0-9.:]+}/{method:ip}/{id:[0-9.:]+}", ApiNeighborHandler)
	r.HandleFunc("/api/{neighbor:[0-9.:]+}/{method:prefixes}", ApiNeighborHandler)
	r.HandleFunc("/api/{method:ipasn}", ApiHandler)

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

	<h3>IP ASN Search</h3>
	<form id="ip_search" class="navbar-search pull-left">
	 	<input type="text" class="search-query" placeholder="IP address" name="ip">
	</form>
	<br>
	<div id="ip_search_results" style="clear:both">
	</div>

	<script src="http://st.pimg.net/cdn/libs/jquery/1.8/jquery.min.js"></script>
	<script src="http://st.pimg.net/cdn/libs/underscore/1/underscore-min.js"></script>
	<script>
		"use strict";
		(function ($) {
			var request;
			$("#ip_search").submit(function(event){
			    if (request) {
			        request.abort();
			    }
			    var $form = $(this);
			    var $inputs = $form.find("input, select, button, textarea");
			    var serializedData = $form.serialize();
			
			    $inputs.prop("disabled", true);

			    var request = $.ajax({
			        url: "/api/ipasn",
			        type: "post",
			        data: serializedData
			    });

			    request.done(function (response, textStatus, jqXHR) {
			    	var result = response.result;
					var html = '<table class="table table-hover" style="width:300px"><tbody>' +
						'<thead><tr><th>Neighbor</th><th>ASN</th></tr></thead>';

			    	_.each(result, function(asn,neighbor) {
			    		html += "<tr><td>" + neighbor + "</td><td>" + asn.ASN + "</td></tr>";
			    	});

			    	html += "</tbody></table>";
			    	$("#ip_search_results").html(html);
			        console.log("Hooray, it worked!", response);
			    });
			
			    request.fail(function (jqXHR, textStatus, errorThrown){
			        console.error("Request error : "+textStatus, errorThrown);
			    });

			    request.always(function () {
			        $inputs.prop("disabled", false);
			    });

			    event.preventDefault();
			});
	})(jQuery);
	</script>

	</body>
</html>
`
