package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	client "github.com/rob121/airfoil-go"
	"github.com/rob121/vhelp"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
	"time"
)

var conf *viper.Viper
var cerr error
var ca *client.AirfoilConn
var ready_to_serve bool = false
var mc mqtt.Client
var debug bool = false

//sample implementation to send states to MQTT on Home assistant

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Starting Server...looking for Airfoil Install")

	flag.BoolVar(&debug, "debug", false, "Debug Flag bool")
	flag.Parse()

	vhelp.Load("config")
	conf, cerr = vhelp.Get("config")

	if cerr != nil {

		log.Fatalf("Unable to load configuration %s", cerr)
	}

	go startHTTPServer()

	mc = mqttClient()

	_, err := client.Scan()

	if err != nil {
		log.Println(err)
	}

	if len(client.Airfoils) < 1 {

		log.Println("No Airfoil Installs Found, check your network settings")
		os.Exit(1)
		return
	}

	//take the first one, this needs some more thinking
	addr := client.Airfoils[0]

	fmt.Printf("Found Airfoil at %s\n", addr)

	ca = client.NewConn(addr)

	derr := ca.Dial()

	if derr != nil {

		log.Println(derr)
		os.Exit(1)
		return
	}

	ready_to_serve = true

	go ca.KeepAlive()
	go fetchData()
	go syncSpeakers()

	//handle messages back from airfoil and do custom actions

	ca.Reader(func(response client.AirfoilResponse, err error) {

		if err != nil {

			log.Printf("Client Error %s\n", err.Error())
			return
		}

		if response.Request == "speakerVolumeChanged" || response.Request == "speakerConnectedChanged" {

			spk, err := ca.GetSpeaker(response.Data.LongIdentifier)
			if err == nil {
				publishPlayerState(spk, mc)
			}

		}

		if response.ReplyID == "13" {

			publishSources()

		}

		if response.ReplyID == "3" {

			for _, spk := range response.Data.Speakers {

				publishMediaPlayer(spk, mc)

			}

		}

	})

	select {}
}

func cleanSpeakerName(spk string) string {

	s := ""

	parts := strings.Split(spk, "@")

	if len(parts) > 1 {
		s = parts[1]
	} else {
		s = spk
	}

	return strings.ToLower(strings.Replace(s, " ", "_", -1)) //strings.Replace("b’", "", strings.Replace(strings.Replace(spk, " ", "_", -1), "'", "", -1), -1))

}

func publishSources() {

	if debug {
		fmt.Println("MQTT Publish Sources State")
	}

	state_topic := fmt.Sprintf("home/speakers/airfoil/sources")

	var out []client.Source

	for _, src := range ca.Sources {

		src.Icon = "" //too much data

		out = append(out, src)
	}

	out2, _ := json.Marshal(out)

	pout, _ := prettyString(string(out2))

	if debug {
		fmt.Println("Sending to topic", state_topic)
		fmt.Println(pout)
	}

	if len(out) > 0 {
		mc.Publish(state_topic, 0, false, string(out2))
	}

	topic2 := fmt.Sprintf("home/speakers/airfoil/source")

	mc.Publish(topic2, 0, false, ca.ActiveSourceKey)

}

func publishPlayerState(spk *client.Speaker, mc mqtt.Client) {

	if debug {
		fmt.Println("MQTT Publish Player State")
	}

	state_topic := fmt.Sprintf("home/speakers/airfoil/%s", cleanSpeakerName(spk.LongIdentifier))

	out := make(map[string]interface{})

	out["id"] = spk.LongIdentifier
	out["friendly_name"] = spk.Name

	if spk.Connected {
		out["connected"] = "on"
	} else {
		out["connected"] = "off"
	}
	out["volume_level"] = spk.Volume

	out2, _ := json.Marshal(out)

	pout, _ := prettyString(string(out2))
	if debug {
		fmt.Println("Sending to topic", state_topic)
		fmt.Println(pout)
	}
	mc.Publish(state_topic, 0, false, string(out2))

}

func publishMediaPlayer(spk client.Speaker, mc mqtt.Client) {

	if debug {
		fmt.Println("MQTT Publish Config")
	}

	//shared state topic with json

	state_topic := fmt.Sprintf("home/speakers/airfoil/%s", cleanSpeakerName(spk.LongIdentifier))

	//publish config for each sensor

	topic := fmt.Sprintf("homeassistant/sensor/airfoil_%s_connected/config", cleanSpeakerName(spk.LongIdentifier))

	out := make(map[string]interface{})

	out["name"] = fmt.Sprintf("airfoil_%s_connected", cleanSpeakerName(spk.LongIdentifier))
	out["unique_id"] = fmt.Sprintf("airfoil_%s_connected", cleanSpeakerName(spk.LongIdentifier))
	out["friendly_name"] = fmt.Sprintf("%s Connected", spk.Name)
	out["state_topic"] = state_topic
	out["value_template"] = "{{ value_json.connected }}"
	out["qos"] = 0
	out["retain"] = false

	outs, _ := json.Marshal(out)

	if debug {
		fmt.Println("Sending to topic", topic)
		pout, _ := prettyString(string(outs))
		fmt.Println(pout)
	}

	mc.Publish(topic, 0, false, string(outs))

	//config for volume

	topic2 := fmt.Sprintf("homeassistant/sensor/airfoil_%s_volume/config", cleanSpeakerName(spk.LongIdentifier))

	out2 := make(map[string]interface{})

	out2["name"] = fmt.Sprintf("airfoil_%s_volume", cleanSpeakerName(spk.LongIdentifier))
	out2["unique_id"] = fmt.Sprintf("airfoil_%s_volume", cleanSpeakerName(spk.LongIdentifier))
	out2["friendly_name"] = fmt.Sprintf("%s Volume", spk.Name)
	out2["state_topic"] = state_topic
	out2["value_template"] = "{{ value_json.volume_level }}"
	out2["qos"] = 0
	out2["retain"] = false

	out2j, _ := json.Marshal(out2)

	mc.Publish(topic2, 0, false, string(out2j))

	topic3 := fmt.Sprintf("homeassistant/sensor/airfoil_%s_id/config", cleanSpeakerName(spk.LongIdentifier))

	out3 := make(map[string]interface{})

	out3["name"] = fmt.Sprintf("airfoil_%s_id", cleanSpeakerName(spk.LongIdentifier))
	out3["unique_id"] = fmt.Sprintf("airfoil_%s_id", cleanSpeakerName(spk.LongIdentifier))
	out3["friendly_name"] = fmt.Sprintf("%s ID", spk.Name)
	out3["state_topic"] = state_topic
	out3["value_template"] = "{{ value_json.id }}"
	out3["qos"] = 0
	out3["retain"] = false

	out3s, _ := json.Marshal(out3)

	mc.Publish(topic3, 0, false, string(out3s))

	topic4 := fmt.Sprintf("homeassistant/sensor/airfoil_sources/config")

	out4 := make(map[string]interface{})

	out4["name"] = fmt.Sprintf("airfoil_sources")
	out4["unique_id"] = fmt.Sprintf("airfoil_sources")
	out4["friendly_name"] = fmt.Sprintf("Airfoil Sources")
	out4["state_topic"] = fmt.Sprintf("home/speakers/airfoil/sources")
	out4["qos"] = 0
	out4["value_template"] = "{{ value_json }}"
	out4["retain"] = false

	out4s, _ := json.Marshal(out4)

	mc.Publish(topic4, 0, false, string(out4s))

	topic5 := fmt.Sprintf("homeassistant/sensor/airfoil_source/config")

	out5 := make(map[string]interface{})

	out5["name"] = fmt.Sprintf("airfoil_source")
	out5["unique_id"] = fmt.Sprintf("airfoil_source")
	out5["friendly_name"] = fmt.Sprintf("Airfoil Source")
	out5["state_topic"] = fmt.Sprintf("home/speakers/airfoil/source")
	out5["qos"] = 0
	out5["retain"] = false

	out5s, _ := json.Marshal(out5)

	mc.Publish(topic5, 0, false, string(out5s))

	go publishPlayerState(&spk, mc)
	go publishSources()

}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("MQTT: Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("MQTT: Connect Lost: %v", err)
}

func mqttClient() mqtt.Client {

	var broker = conf.GetString("mqtt.host")
	var port = conf.GetInt("mqtt.port")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername(conf.GetString("mqtt.user"))
	opts.SetPassword(conf.GetString("mqtt.pass"))
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return client

}

func prettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

// keep the connection alive
func fetchData() {

	go func() {
		for {
			stat := ca.FetchSources() //reload sources occasionally

			if stat != nil {
				log.Printf("Fetch Status %s\n", stat)
				time.Sleep(time.Second * 2)
				continue
			}
			return

		}
	}()

	tm := time.NewTicker(time.Second * 30)

	for range tm.C {

		ca.FetchMetadata()
		ca.FetchSources()

	}

}

func syncSpeakers() {

	tick := time.NewTicker(time.Second * 30)

	for range tick.C {

		ca.SpeakerLock.RLock()
		for _, spk := range ca.Speakers {

			publishMediaPlayer(spk, mc)
			publishPlayerState(&spk, mc)
			publishSources()

		}
		ca.SpeakerLock.RUnlock()

	}

}
