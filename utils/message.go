package utils

type Message struct {
	Header   Header
	Request  map[string]interface{}
	Response map[string]interface{}
}

type Header struct {
	Ttl          int
	Timestamp    string
	ReplyAddress string
	Int_guid     string
	Sender       string
	Receiver     string
	Operation    string
}
