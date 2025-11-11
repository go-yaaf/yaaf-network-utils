package model

type BigQueryRequest struct {
	RequestId   string     `json:"requestId"`
	Caller      string     `json:"caller"`
	SessionUser string     `json:"sessionUser"`
	Calls       [][]string `json:"calls"` // each call is [ip]
}
