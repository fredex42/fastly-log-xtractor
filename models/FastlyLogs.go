package models

import "fmt"

type FastlyLog struct {
	Timestamp        string `json:"timestamp"`
	ClientIp         string `json:"client_ip"`
	GeoCountry       string `json:"geo_country"`
	GeoCity          string `json:"geo_city"`
	Host             string `json:"host"`
	Url              string `json:"url"`
	RequestMethod    string `json:"request_method"`
	RequestProtocol  string `json:"request_protocol"`
	RequestReferrer  string `json:"request_referer"`
	RequestUserAgent string `json:"request_user_agent"`
	ResponseState    string `json:"response_state"`
	ResponseStatus   int16  `json:"response_status"`
	ResponseReason   string `json:"response_reason"`
	ResponseBodySize int64  `json:"response_body_size"`
	FastlyServer     string `json:"fastly_server"`
	FastlyIsEdge     bool   `json:"fastly_is_edge"`
}

func (l *FastlyLog) ToCSV() []string {
	return []string{
		l.Timestamp,
		l.ClientIp,
		l.GeoCountry,
		l.GeoCity,
		l.RequestMethod,
		fmt.Sprintf("%d", l.ResponseStatus),
		l.RequestProtocol,
		l.RequestReferrer,
		l.RequestUserAgent,
		l.ResponseState,
		l.ResponseReason,
		fmt.Sprintf("%d", l.ResponseBodySize),
		l.FastlyServer,
		fmt.Sprintf("%t", l.FastlyIsEdge),
		l.Host,
		l.Url,
	}
}

func CSVHeaderRow() []string {
	return []string{
		"Timestamp",
		"Client IP",
		"Country",
		"City",
		"HTTP Method",
		"HTTP Response Code",
		"HTTP Protocol",
		"Referrer",
		"User Agent",
		"CDN State",
		"Response reason",
		"Response body size",
		"Responding Fastly server",
		"Is on edge?",
		"Host header",
		"Requested URL",
	}
}
