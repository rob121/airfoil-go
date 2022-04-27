package main

import (
    "log"
    "scratch/airfoilapi/client"
    "time"
)

var ca *client.AirfoilConn


func main(){
    
    _,err := client.Scan()

    if(err!=nil){
        log.Println(err)
    }

    log.Println(client.Airfoils)

    if(len(client.Airfoils)<1){

      log.Println("Installs Not Found")
      return
    }

    

    addr := "192.168.20.189:20875"

    go startHTTPServer()

    ca = client.NewConn()

    derr := ca.Dial(addr)

    if(derr!=nil){

        log.Println(derr)
        return
    }

    go watchConn(addr)

    ca.Reader(func(response client.AirfoilResponse,err error){

        if(err!=nil){

            log.Println(err)
            return
        }

        log.Printf("%#v",response)

    })

    select{}
}

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
