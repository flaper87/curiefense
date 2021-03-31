package entities

import ald "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"

type LogEntry struct {
	FullEntry *ald.HTTPAccessLogEntry
	CfLog     CuriefenseLog
	//CurieProxyLog CurieProxyLog
}
