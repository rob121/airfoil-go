package airfoilgo

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
	"sync"
	"time"
)

const PROTOCOL_VERSION = "com.rogueamoeba.protocol.slipstreamremote\nmajorversion=1,minorversion=5\n"
const PROTOCOL_REGEXP = ".*majorversion=1,minorversion=5.*"
const OK_REGEXP = "^OK\n$"

//unimplemented requests

/*

{"request":"getSourceMetadata","requestID":"9","data":{"scaleFactor":1,"requestedData":{"album":true,"remoteControlAvailable":true,"machineIconAndScreenshot":64,"bundleid":true,"albumArt":64,"sourceName":true,"title":true,"icon":16,"trackMetadataAvailable":true,"artist":true,"machineModel":true,"machineName":true}}}
//{"request":"setSpeakerVolume","requestID":"10","data":{"longIdentifier":"542A1B639776@Office","volume":0.60000002384185791}}"

*/

var versioncheck *regexp.Regexp
var okcheck *regexp.Regexp
var readlen = 1024
var Airfoils []string

type AirfoilConn struct {
	Status      int
	Conn        net.Conn
	Cb          func(AirfoilResponse, error)
	Speakers    map[string]Speaker
	Sources     map[string]Source
	SpeakerLock sync.RWMutex
	SourceLock  sync.RWMutex
}

func NewConn() *AirfoilConn {
	versioncheck, _ = regexp.Compile(PROTOCOL_REGEXP)
	okcheck, _ = regexp.Compile(OK_REGEXP)
	conn := &AirfoilConn{}
	conn.Speakers = make(map[string]Speaker)
	conn.Sources = make(map[string]Source)
	return conn
}

func Scan() ([]string, error) {

	resolver, err := zeroconf.NewResolver(nil)

	var s []string

	if err != nil {
		return s, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	defer cancel()

	entries := make(chan *zeroconf.ServiceEntry)

	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {

			addr := fmt.Sprintf("%s:%d", entry.AddrIPv4, entry.Port)

			Airfoils = append(Airfoils, addr)

		}

	}(entries)

	err = resolver.Browse(ctx, "_slipstreamrem._tcp", "local.", entries)

	if err != nil {
		return s, err
	}

	time.Sleep(15 * time.Second)

	return Airfoils, nil
}

func (a *AirfoilConn) Reader(cb func(AirfoilResponse, error)) {
	a.Cb = cb
}

func (a *AirfoilConn) Send(msg string) error {

	if a.Status > 2 {

		return a.send(msg)
	}

	return errors.New("Connection Status Not Ready")
}

func (a *AirfoilConn) send(msg string) error {

	ml := len(msg)
	payload := fmt.Sprintf("%d;%s", ml, msg)
	// log.Printf("Sent:%s\n",payload)
	_, werr := a.Conn.Write([]byte(payload))
	if werr != nil {
		return werr
	}

	return nil
}

func (a *AirfoilConn) Dial(addr string) error {

	var err error

	a.Status = 1
	a.Conn, err = net.Dial("tcp", addr)

	if err != nil {

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

		parts := strings.Split(s, ";")

		if len(parts) > 0 {

			readto, cerr := strconv.Atoi(parts[0])

			if cerr == nil {

				if readto > readlen {

					consumelen := (readto + (len(strconv.Itoa(readto)) + 1)) - readlen //this accouts for the "1234;" length at the beginning of the msg
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

		//log.Printf("Raw String Response: %s\n", s)

		//dont pass up to client til handshake done
		if a.Status > 2 {

			resp, err := a.parse(s)

			a.intercept(resp, err)

			//log.Println("Intercepted: *****", fmt.Sprintf("%+v", resp), " - ", fmt.Sprintf("%+v", err), " *******")

			a.Cb(resp, err)
		}

		if a.Status < 2 && versioncheck.MatchString(s) {

			///log.Println("Got Protocol Request")

			_, cerr := a.Conn.Write([]byte(PROTOCOL_VERSION))

			if cerr != nil {
				log.Println(err)
			}

			a.Status = 2

			time.Sleep(500 * time.Millisecond)
		}

		if a.Status == 2 && okcheck.MatchString(s) {

			//log.Println("Got Ok")

			_, werr := a.Conn.Write([]byte("OK\n"))

			if werr == nil {

				a.Status = 3
				subscribe := `{"request":"subscribe","requestID":"3","data":{"notifications":["remoteControlChangedRequest","speakerConnectedChanged","speakerListChanged","speakerNameChanged","speakerPasswordChanged","speakerVolumeChanged"]}}`
				werr2 := a.Send(subscribe)
				if werr2 != nil {
					log.Println("Subscribed!")
				}
			}

		}

		//fmt.Println("len", binary.Size(buf))
	}
}

// handle syncing states  to the speaker struct
func (a *AirfoilConn) intercept(response AirfoilResponse, err error) {

	if response.Request == "speakerListChanged" {
		a.SpeakerLock.RLock()
		for _, sp := range response.Data.Speakers {
			//updating speaker struct
			a.SetSpeaker(&sp)
		}
		a.SpeakerLock.Unlock()

	}
	//handle sources
	if response.ReplyID == "9" {

		a.SourceLock.Lock()
		for _, sc := range response.Data.Sources {
			a.Sources[sc.Identifier] = sc
		}
		a.SourceLock.Unlock()

	}

	if response.Request == "speakerConnectedChanged" {

		spk, err := a.GetSpeakder(response.Data.LongIdentifier)

		if err == nil {
			spk.Connected = response.Data.Connected
			a.SetSpeaker(spk)
		}

	}

	if response.Request == "speakerVolumeChanged" {

		spk, err := a.GetSpeakder(response.Data.LongIdentifier)

		if err == nil {
			spk.Volume = response.Data.Volume
			a.SetSpeaker(spk)
		}

	}

}

func (a *AirfoilConn) parse(resp string) (AirfoilResponse, error) {

	//log.Printf("Raw Parse: %s\n", resp)

	parts := strings.Split(resp, ";")

	var di AirfoilResponse

	if len(parts) == 2 {

		msglen, _ := strconv.Atoi(parts[0])
		if msglen == len(parts[1]) {
			//got some json decode!

			e := json.Unmarshal([]byte(parts[1]), &di)

			if e != nil {
				return di, e
			}

			if di.ReplyID == "3" {

				for _, s := range di.Data.Speakers {

					a.SetSpeaker(&s)
				}

			} else if di.ReplyID == "9" { //this is a source response

				var sr SourceResponse
				e2 := json.Unmarshal([]byte(parts[1]), &sr)

				var out []Source
				if e2 == nil {

					for typ, items := range sr.Data {

						for _, it := range items {
							it.Type = typ

							out = append(out, it)

						}
					}
					di.Data.Sources = out
				}

			}

			return di, nil
		}
	} else {

		return di, errors.New(fmt.Sprintf("Unparsable Response %s", resp))
	}

	return di, nil

}

func (a *AirfoilConn) Connect(id string) error {

	req := AirfoilRequest{Request: "connectToSpeaker", RequestID: "5", Data: DataRequest{LongIdentifier: id}}

	outb, err := json.Marshal(req)

	if err != nil {

		return err

	}

	return a.Send(string(outb))

}

func (a *AirfoilConn) Volume(id string, vol float64) error {

	req := AirfoilRequest{Request: "setSpeakerVolume", RequestID: "10", Data: DataRequest{LongIdentifier: id, Volume: vol}}

	outb, err := json.Marshal(req)

	if err != nil {

		return err

	}

	return a.Send(string(outb))

}

func (a *AirfoilConn) SetSpeaker(spkr *Speaker) error {

	a.SpeakerLock.Lock()
	a.Speakers[spkr.LongIdentifier] = *spkr
	a.SpeakerLock.Unlock()
	return nil

}

func (a *AirfoilConn) GetSpeakder(id string) (*Speaker, error) {

	var sd *Speaker
	a.SpeakerLock.RLock()
	for _, s := range a.Speakers {

		if id == s.LongIdentifier {
			a.SpeakerLock.RUnlock()
			return &s, nil
		}

	}
	a.SpeakerLock.RUnlock()

	return sd, errors.New("NOT_FOUND")

}

func (a *AirfoilConn) GetSource(id string) (*Source, error) {

	var sd *Source
	a.SourceLock.RLock()
	for _, s := range a.Sources {

		if id == s.Identifier {
			a.SourceLock.RUnlock()
			return &s, nil
		}

	}
	a.SourceLock.RUnlock()

	return sd, errors.New("NOT_FOUND")

}

func (a *AirfoilConn) FetchSources() error {

	req := AirfoilRequest{Request: "getSourceList", RequestID: "9", Data: DataRequest{IconSize: 16, ScaleFactor: 1}}

	outb, err := json.Marshal(req)

	if err != nil {

		return err

	}

	return a.Send(string(outb))

}

func (a *AirfoilConn) SetSource(ident string) error {

	src, e := a.GetSource(ident)

	if e != nil {
		return e
	}

	req := AirfoilRequest{Request: "selectSource", RequestID: "5", Data: DataRequest{Type: src.Type, Identifier: src.Identifier}}

	outb, err := json.Marshal(req)

	if err != nil {

		return err

	}

	return a.Send(string(outb))

}

func (a *AirfoilConn) Disconnect(id string) error {

	req := AirfoilRequest{Request: "disconnectSpeaker", RequestID: "7", Data: DataRequest{LongIdentifier: id}}

	outb, err := json.Marshal(req)

	if err != nil {

		return err

	}

	return a.Send(string(outb))

}
