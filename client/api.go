package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grandcat/zeroconf"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const PROTOCOL_VERSION = "com.rogueamoeba.protocol.slipstreamremote\nmajorversion=1,minorversion=5\n"
const PROTOCOL_REGEXP = ".*majorversion=1,minorversion=5.*"
const OK_REGEXP = "^OK\n$"

var versioncheck *regexp.Regexp
var okcheck *regexp.Regexp
var readlen = 1024
var Airfoils []string

type AirfoilConn struct{ 
  Status int
  Conn net.Conn
  Cb func(AirfoilResponse,error)
  Speakers map[string]Speaker
}

func NewConn() (*AirfoilConn){
	versioncheck, _ = regexp.Compile(PROTOCOL_REGEXP)
	okcheck, _ = regexp.Compile(OK_REGEXP)
    conn := &AirfoilConn{}
    conn.Speakers = make(map[string]Speaker)
    return conn
}

func Scan() ([]string,error){

	resolver, err := zeroconf.NewResolver(nil)

	var s []string

	if err != nil {
		return s,err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	defer cancel()

	entries := make(chan *zeroconf.ServiceEntry)

	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {

			addr := fmt.Sprintf("%s:%d",entry.AddrIPv4,entry.Port)

		    Airfoils=append(Airfoils,addr)

		}
		//log.Println("No more entries.")
	}(entries)

	err = resolver.Browse(ctx, "_slipstreamrem._tcp", "local.", entries)

	if err != nil {
		return s,err
	}

	time.Sleep(15 * time.Second)

	return Airfoils,nil
}

func (a *AirfoilConn) Reader(cb func(AirfoilResponse,error)){
	a.Cb = cb
}

func (a *AirfoilConn) Send(msg string) error{

	if(a.Status>2){

      return a.send(msg)
	}

	return errors.New("Connection Status Not Ready")
}

func (a *AirfoilConn) send(msg string) error{

	ml := len(msg)
	payload:=fmt.Sprintf("%d;%s",ml,msg)
	// log.Printf("Sent:%s\n",payload)
	_,werr := a.Conn.Write([]byte(payload))
	if(werr!=nil) {
		return werr
	}

	return nil
}

func (a *AirfoilConn) Dial(addr string) error{

  var err error

  a.Status = 1
  a.Conn, err = net.Dial("tcp", addr)

  if(err!=nil){
  
    return err
  }

  go a.handleRequest()


 return nil

}


func (a *AirfoilConn) handleRequest() {
	// Buffer that holds incoming information
	buf := make([]byte, readlen)

	for {
		rlen, err := a.Conn.Read(buf)

		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}

		s := string(buf[:rlen])

		//was there a starting stanza?

		parts := strings.Split(s,";")

		if(len(parts)>0){

			readto,cerr := strconv.Atoi(parts[0])

			if(cerr==nil){

               if(readto> readlen){

               	   consumelen := (readto + (len(strconv.Itoa(readto))+1))  - readlen //this accouts for the "1234;" length at the beginning of the msg
                   //todo give this a timeout
               	   for consumelen > 0 {
					   buf2 := make([]byte, consumelen)

					   len2, err2 := a.Conn.Read(buf2)

					   if err2 != nil {
						   fmt.Println("Error reading:", err2.Error())
						   break
					   }

					   consumelen = consumelen - len2

					   s = s + string(buf2[:len2])

				   }
			   }

			}

		}

		//dont pass up to client til handshake done
		if(a.Status>2) {

			resp,err := a.parse(s)

			a.intercept(resp,err)

			a.Cb(resp,err)
		}

		if(a.Status<2 && versioncheck.MatchString(s)){

			log.Println("Got Protocol Request")

			_,cerr := a.Conn.Write([]byte(PROTOCOL_VERSION))

			if cerr != nil {
				log.Println(err)
			}

			a.Status=2

			time.Sleep(500 * time.Millisecond)
		}


		if(a.Status==2 && okcheck.MatchString(s)){

			log.Println("Got Ok")

			_,werr := a.Conn.Write([]byte("OK\n"))

			if(werr==nil){

				a.Status=3
				subscribe := `{"request":"subscribe","requestID":"3","data":{"notifications":["remoteControlChangedRequest","speakerConnectedChanged","speakerListChanged","speakerNameChanged","speakerPasswordChanged","speakerVolumeChanged"]}}`
				werr2 := a.Send(subscribe)
				if(werr2!=nil){
					log.Println("Subscribed!")
				}
			}

		}


		//fmt.Println("len", binary.Size(buf))
	}
}

func (a *AirfoilConn) intercept(resp AirfoilResponse,err error){

	if(resp.Request=="speakerListChanged"){

        for _,sp := range resp.Data.Speakers {
        	 //updating speaker struct
				a.Speakers[sp.LongIdentifier] = sp
		}

	}

	if(resp.Request=="speakerConnectedChanged"){

		for _,sp := range resp.Data.Speakers {
			//updating speaker struct
			if spkr,ok := a.Speakers[sp.LongIdentifier]; ok {

				spkr.Connected = sp.Connected

				a.Speakers[sp.LongIdentifier] = spkr

			}
		}
	}

	if(resp.Request=="speakerVolumeChanged") {

		for _,sp := range resp.Data.Speakers {
			//updating speaker struct
			if spkr,ok := a.Speakers[sp.LongIdentifier]; ok {

				spkr.Volume = sp.Volume

				a.Speakers[sp.LongIdentifier] = spkr

			}
		}


	}

}

func (a *AirfoilConn) parse(resp string) (AirfoilResponse,error){

	parts := strings.Split(resp,";")

	var di AirfoilResponse

	if len(parts) == 2 {

		msglen,_ := strconv.Atoi(parts[0])
		if (msglen == len(parts[1])) {
			//got some json decode!

			e := json.Unmarshal([]byte(parts[1]),&di)

			if(e!=nil){
				return di,e
			}

			if(di.ReplyID=="3"){

				for _,s := range di.Data.Speakers{

					a.Speakers[s.LongIdentifier]=s
				}

			}

			return di,nil
		}
	}else{

		return di,errors.New(fmt.Sprintf("Unparsable Response %s",resp))
	}

	return di,nil

}

func (a *AirfoilConn) Connect(id string) (error){

	req := AirfoilRequest{Request: "connectToSpeaker", RequestID: "5", Data: DataRequest{LongIdentifier: id}}

	outb,err := json.Marshal(req)

	if(err!=nil){

		return err

	}

	return a.Send(string(outb))


}


func (a *AirfoilConn) Disconnect(id string) (error){
                        
        req := AirfoilRequest{Request: "disconnectSpeaker", RequestID: "5", Data: DataRequest{LongIdentifier: id}}
                        
        outb,err := json.Marshal(req)
                        
        if(err!=nil){
               
                return err
                   
        }               
                   
        return a.Send(string(outb))
                                
 
}          

