package airfoilgo

/*
Reference requests in json

{"request":"setSpeakerVolume","requestID":"8","data":{"longIdentifier":"DC9B9CEFC55C@Kitchen","volume":0.46332660317420959}}

{"request":"getSourceMetadata","requestID":"13","data":{"scaleFactor":2,"requestedData":{"album":true,"remoteControlAvailable":true,"machineIconAndScreenshot":64,"bundleid":true,"albumArt":64,"sourceName":true,"title":true,"icon":16,"trackMetadataAvailable":true,"artist":true,"machineModel":true,"machineName":true}}}

{"request":"getSourceList","requestID":"14","data":{"iconSize":16,"scaleFactor":2}}

//change audio source
//{"request":"selectSource","requestID":"5","data":{"type":"recentApplications","identifier":"\/Applications\/Spotify.app"}}

//connect to speaker
//{"request":"connectToSpeaker","requestID":"5","data":{"longIdentifier":"843835649D9C@Seim's Lappi"}}

//Event for metadata changed
//45;{"request":"sourceMetadataChanged","data":{}}

//fetch metadata

	{
	  "request": "getSourceMetadata",
	  "requestID": "7",
	  "data": {
	    "scaleFactor": 1,
	    "requestedData": {
	      "album": true,
	      "remoteControlAvailable": true,
	      "machineIconAndScreenshot": 64,
	      "bundleid": true,
	      "albumArt": 64,
	      "sourceName": true,
	      "title": true,
	      "icon": 16,
	      "trackMetadataAvailable": true,
	      "artist": true,
	      "machineModel": true,
	      "machineName": true
	    }
	  }
	}
*/
type AirfoilResponse struct {
	ReplyID string       `json:"replyID"`
	Data    DataResponse `json:"data"`
	Request string       `json:"request"`
}

type DataResponse struct {
	Speakers         []Speaker `json:"speakers"`
	Sources          []Source  `json:"sources,omitempty"`
	CanRemoteControl bool      `json:"canRemoteControl"`
	CanConnect       bool      `json:"canConnect"`
	Notifications    []string  `json:"notifications"`
	LongIdentifier   string    `json:"longIdentifier"`
	Connected        bool      `json:"connected,omitempty"`
	Volume           float64   `json:"volume,omitempty"`
}

type Speaker struct {
	Password       bool    `json:"password"`
	Volume         float64 `json:"volume"`
	LongIdentifier string  `json:"longIdentifier"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Connected      bool    `json:"connected"`
}

type SourceResponse struct {
	ReplyID string              `json:"replyID"`
	Data    map[string][]Source `json:"data"`
}

type Source struct {
	FriendlyName string `json:"friendlyName"`
	Icon         string `json:"icon"`
	Identifier   string `json:"identifier"`
	Type         string `json:"type"`
}

type AirfoilRequest struct {
	Request   string      `json:"request"`
	RequestID string      `json:"requestID"`
	Data      DataRequest `json:"data"`
}

type DataRequest struct {
	Type           string  `json:"type"`
	Identifier     string  `json:"identifier"`
	LongIdentifier string  `json:"longIdentifier"`
	ScaleFactor    int     `json:"scaleFactor"`
	IconSize       int     `json:"iconSize"`
	Volume         float64 `json:"volume"`
}
