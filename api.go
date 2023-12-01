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

var versioncheck *regexp.Regexp
var okcheck *regexp.Regexp
var readlen = 1024
var maxbuffer = 16384
var Airfoils []string

type AirfoilConn struct {
	Status      int
	Address     string
	Conn        net.Conn
	Cb          func(AirfoilResponse, error)
	Speakers    map[string]Speaker
	Sources     map[string]Source
	SpeakerLock sync.RWMutex
	SourceLock  sync.RWMutex
	Errors      []error
}

func NewConn(addr string) *AirfoilConn {
	versioncheck, _ = regexp.Compile(PROTOCOL_REGEXP)
	okcheck, _ = regexp.Compile(OK_REGEXP)
	conn := &AirfoilConn{}
	conn.Speakers = make(map[string]Speaker)
	conn.Sources = make(map[string]Source)
	conn.Address = addr
	return conn
}

func Scan() ([]string, error) {

	resolver, err := zeroconf.NewResolver(nil)

	var s []string

	Airfoils = nil

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

	a.Conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	ml := len(msg)
	payload := fmt.Sprintf("%d;%s", ml, msg)
	log.Printf("Sending Request:%s\n", payload)
	_, werr := a.Conn.Write([]byte(payload))
	if werr != nil {
		return werr
	}

	return nil
}

func (a *AirfoilConn) Close() error {

	if a.Conn != nil {
		//close if an existing connection

		return a.Conn.Close()

	}

	return nil

}

func (a *AirfoilConn) KeepAlive() {

	tick := time.NewTicker(time.Second * 10)

	for range tick.C {

		stat := a.Ping()

		if stat != nil {

			res, _ := Scan()

			if len(res) > 0 {
				a.Address = res[0]
			}

			a.Status = 0 //reset and redial!

			a.Dial()
		}

	}

}

func (a *AirfoilConn) Ping() error {

	d := net.Dialer{Timeout: time.Second * 5}
	_, err := d.Dial("tcp", a.Address)

	return err

}

func (a *AirfoilConn) Dial() error {

	var err error

	a.Status = 1

	a.Close() //close existing

	d := net.Dialer{Timeout: time.Second * 5}
	a.Conn, err = d.Dial("tcp", a.Address)

	if err != nil {

		return err
	}

	go a.handleRequest()

	return nil

}

func (a *AirfoilConn) handleRequest() {
	// Buffer that holds incoming information
	buf := make([]byte, readlen)

	re := regexp.MustCompile("[0-9]+;")

	for {
		rlen, err := a.Conn.Read(buf)

		if err != nil {
			fmt.Println("Read Error:", err.Error())
			return //close it down
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
							fmt.Println("Read Error:", err2.Error())
							return //close it down
						}

						consumelen = consumelen - len2

						s = s + string(buf2[:len2])

					}
				}

			}

		}

		log.Printf("Raw String Response: %s\n", s)

		//dont pass up to client til handshake done
		if a.Status > 2 {

			//we may have gotten multiple segments in one string, so break it up

			split := re.Split(s, -1)

			for i := range split {

				if len(split[i]) > 0 {
					//because we are handling json with a leading bit of data and we have to split on it to detect, lets put something back

					go func(sp string) {
						resp, serr := a.parse(fmt.Sprintf("%d;%s", len(sp), sp))

						a.intercept(resp, serr)

						a.Cb(resp, serr)
					}(split[i])
				}

			}

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

	if response.Request == "speakerListChanged" || response.ReplyID == "3" {

		for _, sp := range response.Data.Speakers {
			//updating speaker struct
			a.SetSpeaker(&sp)
		}

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

		spk, err := a.GetSpeaker(response.Data.LongIdentifier)

		if err == nil {
			spk.Connected = response.Data.Connected
			a.SetSpeaker(spk)
		}

	}

	if response.Request == "speakerVolumeChanged" {

		spk, err := a.GetSpeaker(response.Data.LongIdentifier)

		if err == nil {
			spk.Volume = response.Data.Volume
			a.SetSpeaker(spk)
		}

	}

}

func (a *AirfoilConn) Subscribe() error {

	a.Status = 3
	subscribe := `{"request":"subscribe","requestID":"3","data":{"notifications":["remoteControlChangedRequest","speakerConnectedChanged","speakerListChanged","speakerNameChanged","speakerPasswordChanged","speakerVolumeChanged"]}}`
	return a.Send(subscribe)

}

func (a *AirfoilConn) parse(resp string) (AirfoilResponse, error) {

	log.Printf("String to Parse: %s\n", resp)

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

func (a *AirfoilConn) GetSpeaker(id string) (*Speaker, error) {

	var sd *Speaker
	a.SpeakerLock.RLock()
	defer a.SpeakerLock.RUnlock()
	for _, s := range a.Speakers {

		if id == s.LongIdentifier {

			return &s, nil
		}

	}

	return sd, errors.New("NOT_FOUND")

}

func (a *AirfoilConn) GetSource(id string) (*Source, error) {

	var sd *Source
	a.SourceLock.RLock()
	defer a.SourceLock.RUnlock()
	for _, s := range a.Sources {

		if id == s.Identifier {

			return &s, nil
		}

	}

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
