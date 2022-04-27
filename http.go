package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"scratch/airfoilapi/client"
	"time"
)

type JsonResp struct {
	Code    int         `json:"code"`
	Payload interface{} `json:"payload"`
	Message string      `json:"message"`
}

func startHTTPServer(){

	r:= mux.NewRouter()
	r.HandleFunc("/", httpDefaultHandler)
	r.HandleFunc("/airfoils", httpAirfoilsHandler)
	r.HandleFunc("/connect/{id}", httpConnectHandler)
	r.HandleFunc("/disconnect/{id}", httpDisconnectHandler)
	r.HandleFunc("/speakers", httpSpeakersHandler)
	http.Handle("/", r)

	page_port:="8086"

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf(":%s", page_port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Listening on port", page_port)
	log.Fatal(srv.ListenAndServe())
}

func httpConnectHandler(w http.ResponseWriter, r *http.Request){

	vars := mux.Vars(r)

	id := vars["id"]

	if(len(id)<1){

		respond(w,500,"Error","Invalid Id")
		return

	}

	for _,s := range ca.Speakers {

		if(id==s.LongIdentifier) {
			ca.Connect(id)
			respond(w, 200, "OK", "")
			return
		}

	}

	respond(w,500,"Error","Id Mismatch")
	return

}

func httpDisconnectHandler(w http.ResponseWriter, r *http.Request){
        
        vars := mux.Vars(r)

        id := vars["id"]

        if(len(id)<1){
        
                respond(w,500,"Error","Invalid Id")   
                return
        
        }
        
        for _,s := range ca.Speakers {
        
                if(id==s.LongIdentifier) {
                        ca.Disconnect(id)
                        respond(w, 200, "OK", "")
                        return
                }
                
        }
         
        respond(w,500,"Error","Id Mismatch")
        return
        
}

func httpDefaultHandler(w http.ResponseWriter, r *http.Request) {
	respond(w,200,"OK","")
}

func httpAirfoilsHandler(w http.ResponseWriter, r *http.Request) {
	respond(w,200,"OK",client.Airfoils)
}

func httpSpeakersHandler(w http.ResponseWriter, r *http.Request) {
	respond(w,200,"OK",ca.Speakers)
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
