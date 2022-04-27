# airfoil-go
Golang Api / Server to interact with Airfoil 

## Overview

Airfoil is a mac based software to manage multiple airplay and bluetooth speakers. Airfoil-Go is a library and a api server to manage the connection of those speakers

## Operating Principle

Airfoil-Go emulates a client by using the zeroconf protocl to discover an airfoil instance on the network and then send it commands.

### Current Support
  * Find Airfoil install on network
  * List speakers on Airfoil
  * Connect to speakers on Airfoil
  * Disconnect a connected speaker

### Server

A reference api implementation is available in the cmd folder. Supports the following commands

GET /airfoils
```
{
    "code": 200,
    "payload": [
    "[192.168.1.5]:20875"
    ],
    "message": "OK"
}
```

GET /speakers
```
{
   "code":200,
   "payload":{
      "DC9B9CEFC55C@Kitchen":{
         "password":false,
         "volume":0.4760432839393616,
         "longIdentifier":"DC9B9CEFC55C@Kitchen",
         "name":"Kitchen",
         "type":"airplay",
         "connected":true
      },
      "com.rogueamoeba.airfoil.LocalSpeaker":{
         "password":false,
         "volume":1,
         "longIdentifier":"com.rogueamoeba.airfoil.LocalSpeaker",
         "name":"Computer",
         "type":"local",
         "connected":false
      }
   },
   "message":"OK"
}
```
GET /connect/{longIdentifier}
```
{
"code": 200,
"payload": "",
"message": "OK"
}
```
GET /disconnect/{longIdentifier}
```
{
"code": 200,
"payload": "",
"message": "OK"
}
```

## Credits

Took inspiration from https://github.com/dersimn/Airfoil-Slipstream-Remote-Protocol
