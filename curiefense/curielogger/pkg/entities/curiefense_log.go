package entities

type CuriefenseLog struct {
	RequestId string `json:"requestid"`
	Timestamp string `json:"timestamp"`
	Scheme    string `json:"scheme"`
	Authority string `json:"authority"`
	Port      uint32 `json:"port"`
	Method    string `json:"method"`
	Path      string `json:"path"`

	Blocked     bool                   `json:"blocked"`
	BlockReason map[string]interface{} `json:"block_reason"`
	Tags        []string               `json:"tags"`

	RXTimers RXTimer `json:"rx_timers"`
	TXTimers TXTimer `json:"tx_timers"`

	Upstream   Upstream   `json:"upstream"`
	Downstream Downstream `json:"downstream"`

	TLS      TLS      `json:"tls"`
	Request  Request  `json:"request"`
	Response Response `json:"response"`
	Metadata Metadata `json:"metadata"`
}
