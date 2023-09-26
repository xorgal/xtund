// File: server/api.go
package server

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtun-core/pkg/counter"
	"github.com/xorgal/xtund/internal"
)

func initAPIRoutes(config config.Config, allocator *internal.Allocator) {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		currentTime := time.Now().Unix()
		response := DefaultResponse{
			Timestamp: currentTime,
		}
		sendJsonResponse(w, http.StatusOK, response)
	})

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		if !checkPermission(w, r, config) {
			return
		}
		response := ServerConfigurationResponse{
			BufferSize: config.BufferSize,
			MTU:        config.MTU,
			Compress:   config.Compress,
		}
		sendJsonResponse(w, http.StatusOK, response)
	})

	http.HandleFunc("/allocator/register", func(w http.ResponseWriter, r *http.Request) {
		if !checkPermission(w, r, config) {
			return
		}
		var request RegisterDeviceRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		client, serverIP, err := allocator.RegisterDevice(request.DeviceId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			internal.PrintErr("allocator.RegisterDevice(request.DeviceId):", err)
		}
		response := RegisterDeviceResponse{
			Client: client,
			Server: serverIP,
		}
		sendJsonResponse(w, http.StatusOK, response)
	})

	// Todo: convert to json
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		if !checkPermission(w, r, config) {
			return
		}
		io.WriteString(w, counter.PrintBytes(true))
	})
}

func sendJsonResponse(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
