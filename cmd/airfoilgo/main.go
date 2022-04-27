package main

import (
	"flag"
	client "github.com/rob121/airfoil-go"
	"log"
	"time"
	"fmt"
	"os"
)

var ca *client.AirfoilConn
var http_port string

func main(){

	flag.StringVar(&http_port,"port","8086","Server Api Port")
    flag.Parse()

	fmt.Println("Starting Server...looking for Airfoil Install")

	_,err := client.Scan()

	if(err!=nil){
		fmt.Println(err)
	}

	if(len(client.Airfoils)<1){

		fmt.Println("No Airfoil Installs Found, check your network settings")
		os.Exit(1)
		return
	}

    //take the first one
	addr := client.Airfoils[0]

	fmt.Printf("Found airfoil at %s\n",addr)

	go startHTTPServer()

	ca = client.NewConn()

	derr := ca.Dial(addr)

	if(derr!=nil){

		fmt.Println(derr)
		os.Exit(1)
		return
	}

	go watchConn(addr)

	ca.Reader(func(response client.AirfoilResponse,err error){

		if(err!=nil){

			fmt.Println(err)
			return
		}

		fmt.Printf("Response: \n%#v\n",response)

	})

	select{}
}
//keep the connection alive
func watchConn(addr string){

	tick := time.NewTicker(time.Second * 300)

	for range tick.C {

		derr := ca.Dial(addr)

		if(derr!=nil){

			log.Println(derr)
			return
		}
	}
}