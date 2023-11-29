package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	client "github.com/rob121/airfoil-go"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type JsonResp struct {
	Code    int         `json:"code"`
	Payload interface{} `json:"payload"`
	Message string      `json:"message"`
}

func startHTTPServer() {

	r := mux.NewRouter()
	r.Use(Middleware)
	r.HandleFunc("/", httpDefaultHandler)
	r.HandleFunc("/airfoils", httpAirfoilsHandler)
	r.HandleFunc("/connect/{id}", httpConnectHandler)
	r.HandleFunc("/toggleconn/{id}", httpToggleconnHandler)
	r.HandleFunc("/source/{id}", httpSourceHandler)
	r.HandleFunc("/volume/{id}/{vol}", httpVolumeHandler)
	r.HandleFunc("/disconnect/{id}", httpDisconnectHandler)
	r.HandleFunc("/speakers", httpSpeakersHandler)
	r.HandleFunc("/sources", httpSourcesHandler)
	http.Handle("/", r)

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf(":%s", conf.GetString("port")),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Listening on port", conf.GetString("port"))
	log.Fatal(srv.ListenAndServe())
}

func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !ready_to_serve {
			respond(w, 500, "Error", "Not Ready")
			return
		}

		h.ServeHTTP(w, r)
	})
}

func httpVolumeHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id := vars["id"]

	vol := vars["vol"]

	if strings.Contains(vol, ".") {

		parts := strings.Split(vol, ".")
		vol = parts[0]

	}

	voli, _ := strconv.Atoi(vol)

	log.Printf("Got Volume %s for %s", vol, id)

	if len(id) < 1 {

		respond(w, 500, "Error", "Invalid Id")
		return

	}

	if len(vol) < 1 || voli > 100 || voli < 1 {
		voli = 1
	}

	spk, err := ca.GetSpeakder(id)

	if err != nil {
		respond(w, 500, "Error", err.Error())
		return
	}

	volf := float64(voli) / 100

	status := ca.Volume(spk.LongIdentifier, volf)

	if status != nil {
		respond(w, 500, "Error", status.Error())
		return
	}

	respond(w, 200, "OK", "")

}

func httpConnectHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id := vars["id"]

	if len(id) < 1 {

		respond(w, 500, "Error", "Invalid Id")
		return

	}

	for _, s := range ca.Speakers {

		if id == s.LongIdentifier {
			resp := ca.Connect(id)

			if resp == nil {
				respond(w, 200, "OK", "")
			} else {

				respond(w, 500, fmt.Sprintf("ERR: %s", resp), "")
			}

			return
		}

	}

	respond(w, 500, "Error", "Id Mismatch")
	return

}

func httpToggleconnHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id := vars["id"]

	if len(id) < 1 {

		respond(w, 500, "Error", "Invalid Id")
		return

	}

	spk, err := ca.GetSpeakder(id)

	var resp error

	if err == nil {

		if spk.Connected == true {

			resp = ca.Disconnect(id)

		} else {

			resp = ca.Connect(id)

		}

		if resp == nil {
			respond(w, 200, "OK", "")
			return
		} else {

			respond(w, 500, fmt.Sprintf("ERR: %s", resp), "")
			return
		}

	}

	respond(w, 500, "Error", "Id Mismatch")
	return

}

func httpSourceHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id := vars["id"]

	if len(id) < 1 {

		respond(w, 500, fmt.Sprintf("ERR:Id Not Found"), "")

	}

	resp := ca.SetSource(id)

	respond(w, 200, "OK", resp)

}

func httpDisconnectHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id := vars["id"]

	if len(id) < 1 {

		respond(w, 500, "Error", "Invalid Id")
		return

	}

	for _, s := range ca.Speakers {

		if id == s.LongIdentifier {
			resp := ca.Disconnect(id)

			if resp == nil {

				respond(w, 200, "OK", "")

			} else {

				respond(w, 500, fmt.Sprintf("ERR: %s", resp), "")
			}
			return
		}

	}

	respond(w, 500, "Error", "Id Mismatch")
	return

}

func httpDefaultHandler(w http.ResponseWriter, r *http.Request) {
	respond(w, 200, "OK", ca)
}

func httpAirfoilsHandler(w http.ResponseWriter, r *http.Request) {
	respond(w, 200, "OK", client.Airfoils)
}

func httpSourcesHandler(w http.ResponseWriter, r *http.Request) {

	ca.FetchSources()
	respond(w, 200, "OK", ca.Sources)

}

func httpSpeakersHandler(w http.ResponseWriter, r *http.Request) {

	respond(w, 200, "OK", ca.Speakers)

}

func respond(w http.ResponseWriter, code int, message string, payload interface{}) {

	resp := JsonResp{
		Code:    code,
		Payload: payload,
		Message: message,
	}

	var jsonData []byte
	jsonData, err := json.Marshal(resp)

	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-type", "application/json")

	fmt.Fprintln(w, string(jsonData))

}
