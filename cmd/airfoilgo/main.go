package main

import (
	"encoding/json"
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

func main() {

	fmt.Println("Starting Server...looking for Airfoil Install")

	vhelp.Load("config")
	conf, cerr = vhelp.Get("config")

	if cerr != nil {

		log.Fatalf("Unable to load configuration %s", cerr)
	}

	go startHTTPServer()

	mc = mqttClient()

	_, err := client.Scan()

	if err != nil {
		fmt.Println(err)
	}

	if len(client.Airfoils) < 1 {

		fmt.Println("No Airfoil Installs Found, check your network settings")
		os.Exit(1)
		return
	}

	ready_to_serve = true

	//take the first one, this needs some more thinking
	addr := client.Airfoils[0]

	fmt.Printf("Found airfoil at %s\n", addr)

	ca = client.NewConn()

	derr := ca.Dial(addr)

	if derr != nil {

		fmt.Println(derr)
		os.Exit(1)
		return
	}

	go watchConn(addr)
	go syncSpeakers()

	ca.Reader(func(response client.AirfoilResponse, err error) {

		if err != nil {

			fmt.Printf("Client Error %s\n", err.Error())
			return
		}

		if response.Request == "speakerVolumeChanged" || response.Request == "speakerConnectedChanged" {

			spk, err := ca.GetSpeakder(response.Data.LongIdentifier)
			if err == nil {
				publishPlayerState(spk, mc)
			}

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

func publishPlayerState(spk *client.Speaker, mc mqtt.Client) {

	fmt.Println("MQTT Publish Player State")

	state_topic := fmt.Sprintf("home/speakers/airfoil/%s", cleanSpeakerName(spk.LongIdentifier))

	out := make(map[string]interface{})

	out["id"] = spk.LongIdentifier
	out["friendly_name"] = spk.Name

	if spk.Connected {
		out["connected"] = "yes"
	} else {
		out["connected"] = "no"
	}
	out["volume_level"] = fmt.Sprint(spk.Volume)

	out2, _ := json.Marshal(out)

	fmt.Println("Sending to topic", state_topic, string(out2))

	tok := mc.Publish(state_topic, 0, false, string(out2))

	fmt.Println(tok)

}

func publishMediaPlayer(spk client.Speaker, mc mqtt.Client) {

	fmt.Println("MQTT Publish Config")

	topic := fmt.Sprintf("homeassistant/sensor/airfoil_%s/config", cleanSpeakerName(spk.LongIdentifier))

	state_topic := fmt.Sprintf("home/speakers/airfoil/%s", cleanSpeakerName(spk.LongIdentifier))

	out := make(map[string]interface{})

	out["name"] = fmt.Sprintf("airfoil_%s", cleanSpeakerName(spk.LongIdentifier))
	out["unique_id"] = fmt.Sprintf("airfoil_%s", cleanSpeakerName(spk.LongIdentifier))
	out["friendly_name"] = fmt.Sprintf(spk.Name)
	out["state_topic"] = state_topic
	out["payload_on"] = "ON"
	out["payload_off"] = "OFF"
	out["state_on"] = "ON"
	out["state_off"] = "OFF"
	out["qos"] = 0
	out["retain"] = true

	go publishPlayerState(&spk, mc)

	out2, _ := json.Marshal(out)

	fmt.Println("Sending to topic", topic, string(out2))

	tok := mc.Publish(topic, 0, false, string(out2))

	fmt.Println(tok)

}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
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

// keep the connection alive
func watchConn(addr string) {

	tick := time.NewTicker(time.Second * 300)

	ca.FetchSources()

	for range tick.C {

		derr := ca.Dial(addr)

		if derr != nil {

			ca.FetchSources() //reload sources occasionally

			log.Println(derr)
			return
		}
	}
}

func syncSpeakers() {

	tick := time.NewTicker(time.Second * 10)

	for range tick.C {

		ca.SpeakerLock.RLock()
		for _, spk := range ca.Speakers {

			publishMediaPlayer(spk, mc)
			publishPlayerState(&spk, mc)

		}
		ca.SpeakerLock.RUnlock()

	}

}
