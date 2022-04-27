package airfoilgo


/*
Reference requests in json

//change audio source
//{"request":"selectSource","requestID":"5","data":{"type":"recentApplications","identifier":"\/Applications\/Spotify.app"}}

//connect to speaker
//{"request":"connectToSpeaker","requestID":"5","data":{"longIdentifier":"843835649D9C@Seim's Lappi"}}

*/
type AirfoilResponse struct {
	ReplyID string       `json:"replyID"`
	Data    DataResponse `json:"data"`
	Request string       `json:"request"`
}

type DataResponse struct {
	Speakers []Speaker     `json:"speakers"`
	CanRemoteControl bool  `json:"canRemoteControl"`
    CanConnect bool        `json:"canConnect"`
	Notifications []string `json:"notifications"`
	LongIdentifier string  `json:"longIdentifier"`
	Connected bool         `json:"connected"`
}

type Speaker struct{
 Password bool `json:"password"`
 Volume float64 `json:"volume"`
 LongIdentifier string `json:"longIdentifier"`
 Name string `json:"name"`
 Type string `json:"type"`
 Connected bool `json:"connected"`
}

type AirfoilRequest struct{
	Request   string      `json:"request"`
	RequestID string      `json:"requestID"`
	Data      DataRequest `json:"data"`
}

type DataRequest struct {
	Type string `json:"type"`
	Identifier string `json:"identifier"`
	LongIdentifier string `json:"longIdentifier"`
}

