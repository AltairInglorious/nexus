package transport

type NATSError struct {
	Status int    `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

type NATSOk struct {
	Status int `json:"status,omitempty"`
	Body   any `json:"body,omitempty"`
}
