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
#### List Airfoils on network
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
#### Fetch Speakers
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
#### Connect to a Speaker
GET /connect/{longIdentifier}
```
{
"code": 200,
"payload": "",
"message": "OK"
}
```



#### Disconnect Speaker
GET /disconnect/{longIdentifier}
```
{
"code": 200,
"payload": "",
"message": "OK"
}
```

#### Change Volume on Speaker
GET /volume/{longIdentifier}/{num}

Where {num} is 1-100 representing the volume

```
{
"code": 200,
"payload": "",
"message": "OK"
}
```

#### Fetch Sources
GET /sources
```
{
"code": 200,
"payload": {
    "/System/Applications/QuickTime Player.app": {
        "friendlyName": "QuickTime Player",
        "icon": "iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAAXNSR0IArs4c6QAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAEKADAAQAAAABAAAAEAAAAAAXnVPIAAADCklEQVQ4EW1TW0sUYRh+vvlGZw+zo+5k22qSaQeKLEMtO0plFARRdLqoLoLoougq6LKLoH/QRV1EIAXRRTcZFUhCWaamUaRJrq4pa7itmbutznyzO9P72YGiXniYZ96Z53nf7/Ay/AiVHjrBIBQSFML/wqOkIGR+wmFEpDiyefuOiyUl5iHO+ULGWAHl/hduPpdLzaS/Pujp7LhsWVaC019G45btl0yz9DwDM1y4XNP8WL22DitrGrB0+SoYRhGmv6TgCJt5YEFN860PLygNj3+Mt8vqRiBoHBRODoz6qa1vRNOeYyhfHIGu+6FSLp11kJxMou3+bfT1dM535vcF9hK5JA00z81HHMdD/cZN2L3/JFZWl2HwwyjudiVAXWHnpnKsWbEEgSNn4DgOertfUDFmkrZQGjDhOKrfH8TW3YdRVl6Kq3c60fKsEIyX0GfgxssUTjV9wplDjWjedxRv+nowO5uVy2fSALZto6a2HuFwMd4NDKFlIgylpghGyAePVp1JC9xMTKN+YBCV0Qiqlq9AX9ePpcwfl7BslEaimMtO41F/HLzahF4VQuuJYtw7FqKNVKBUmmh9G0c+n8GiaHS+qCw+b2Dbc/BcC4mxEczmssACP9yQjfGZDMR0ElfWMZjFc0hNTWA0FoPKPTKwpP6XgYWxkRiy39JYhixYmMEKKjjXPYyhmSysRBxnC0awjq7aaHwMw0Mx2NYfBoJeXj7vgBCzMFQXFR864OkMXyvKcWHoM+IZ6orC+jKJb+kUOp62/+5AbqJnCztvT9m87WErNm7biQOZcSQfv8eIYcLKpBHzHJgaRygUoJORbTMqZstr7cmj0DlXjruuq4+NxuHlBJZUV6M46EcZ8coChhI9gMlPEwjoOrLCQ6FPw2D/2ylh29dkBxlHiPuMKaeJ42n7E7x+9Qp1GxoQiZZBlkmSuLe7C5ovgK27msFp64Ul2qRWauTgVHBFuc5VNaGqao5z1VMJ9O5JTgNGnMBVl3JJrvBbpKmSWjmNMqSJHOcQQSP8yhP9K/4Z5+93wTg69x2l8gAAAABJRU5ErkJggg==",
        "identifier": "/System/Applications/QuickTime Player.app",
        "type": "recentApplications"
    },
    "AppleUSBAudioEngine:C-Media Electronics Inc.:USB Audio Device:100000:2,1": {
        "friendlyName": "USB Audio Device",
        "icon": "iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAAXNSR0IArs4c6QAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAEKADAAQAAAABAAAAEAAAAAAXnVPIAAAC/ElEQVQ4EV1TTUwTQRR+O7t02yIqbQ2giBKKP1HQmKCIJsVAxOrJRA7e8ETiwUg4eQNOXIjGRAJiNd40JKYHhURDLAlQEkOCf3AAb7S2SEsL7ZZtd2d9bxqMYSazmZ35vu/9jgR7Rm9vr6e1tfWcqqouRVEkujYMw9J1PRkKhb4ODQ1t7KEUf/1+v7qwsNATiUSXY7GYkU5vWZqWEyu9tWX9xrO1SGSZMITdFZFp09XVZe/vH3h2tObYIyYxD7c4y2oaLC0tQTQaBbvdAYwpzGazeVxud0fL5atVmpb5uLi4aAgX5+a/PKz31j3e0XfA4haCJZicnITp6WlhyOfzwQ3/TbAsDpIkgV21w8rqr56W5qYncl9fn+dSc8swxuvBWAUhX8hDMBiEXC6HJAvS6TQ0NV1EYRk4GpDQAOJr3a7yN0oDDryo0/W8AJMFApGYxbkQNEwTOArhscAQlklyHXEVE2S3TbUpeTykUUDrFk6OZBOJNEiIREy+Q5ZBwmlTVcFllmEwFC9aNTm8DIxBMrkpSCRAi8SSySQEXozhPwpjLpBC3jBGNsQh+kduZjIZSKVS4HSWQj6vi0X71GYKspltEQZFZpqWEGecS9xAVSwdJkcGLCXMh2eh/bof98fFau/wQzg8J/a7iRR5QS7TuZHQNM3A3AJHd33X2iAei0N4bgY6/LfECs/OwPp6HO/ahbfkPnHMvJGQK2uPZc6cPnvb4XB4CgUDHE4nNDSeh2hkDT5PfYKfP75DNXp1p/MulJXtF8lUlBLYzmZWQlMTg+w59nZ8PT5CYWCXAFVj4n0Qk5YoVkIkMAEf8EzPY6UQQ+6jlyP0LkQrO0rkxeoabxVauMBkWfTAgYPlUH/yFHjrT4DL5YaKiiqoPHwEMy9BZC0SePf29cC/VsaQwIsPZLD7wf1yt6e7dF+Z1+l0yBSrGLjJaTkzm91e3UxsjAZGnw5jq+t0J95CEVX8dnbeO+Rru9JoSbIL6y/uGWP4RMzkzNTst/HxV3/+x/8Fg3q6F3VhtigAAAAASUVORK5CYII=",
        "identifier": "AppleUSBAudioEngine:C-Media Electronics Inc.:USB Audio Device:100000:2,1",
        "type": "audioDevices"
    },
    "com.rogueamoeba.source.systemaudio": {
        "friendlyName": "System Audio",
        "icon": "iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAAXNSR0IArs4c6QAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAEKADAAQAAAABAAAAEAAAAAAXnVPIAAABMElEQVQ4Ee2QwUrDQBCG/91sujSmB6GIFapE1IfoQQTBk0/hw+kD6KXiwYOo16ZgxIJIUSPWWgq2mk02cSZQ+gC9dpZlYGb+b5gfWMbCDggmFEVx0O/HJ1+j4ZqUQjhUk66GrlRgc4ssMUjzjCeR5yhqK7Xh7s7WmRDiQpF4r315dWqMaVABUkporRFFEcKwQ3Cg1Wqh2dyEMQmstQTJ0es9HZL2WLx/fLYfo+iIhbPvOA7G4zFGo2/aKlCv1+H7PrIsK8UMYNB2EFyLm9u758b6RjCdTmh4HgxRSpWFNE1L4axLm0tgHMcvyqXneVVMJj/sxWym3EBngc/i4Mx9zvyrnkcLXFd1w865quj9+O119Tf507BsFb05i45gApvNIJBQWpOYQTd8uOfWMhZ04B87l40eZFPvUwAAAABJRU5ErkJggg==",
        "identifier": "com.rogueamoeba.source.systemaudio",
        "type": "systemAudio"
    }
}
```
#### Sets a source
GET /source/{identifier}
```
{
"code": 200,
"payload": null,
"message": "OK"
}
```

### Todo

- Handle More Than one Airfoil on the network

## Credits

Took inspiration from https://github.com/dersimn/Airfoil-Slipstream-Remote-Protocol
