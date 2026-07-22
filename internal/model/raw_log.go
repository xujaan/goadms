package model

import "time"

type RawLog struct {
	ID            int64     `json:"id"`
	DeviceSN      string    `json:"device_sn,omitempty"`
	RequestMethod string    `json:"request_method,omitempty"`
	RequestURI    string    `json:"request_uri,omitempty"`
	QueryParams   []byte    `json:"query_params,omitempty"`
	RequestBody   string    `json:"request_body,omitempty"`
	ResponseBody  string    `json:"response_body,omitempty"`
	LogType       string    `json:"log_type"`
	CreatedAt     time.Time `json:"created_at"`
}
