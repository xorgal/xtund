// File: server/types.go
package server

type DefaultResponse struct {
	Timestamp int64 `json:"timestamp"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type ServerConfigurationResponse struct {
	BufferSize int  `json:"bufferSize"`
	MTU        int  `json:"mtu"`
	Compress   bool `json:"compress"`
}

type RegisterDeviceRequest struct {
	DeviceId string `json:"id"`
}

type RegisterDeviceResponse struct {
	Server string `json:"server"`
	Client string `json:"client"`
}
