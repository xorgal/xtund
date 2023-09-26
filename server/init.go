// File: server/init.go
package server

import (
	"log"
	"net/http"

	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtun-core/pkg/tun"
	"github.com/xorgal/xtund/internal"
)

func InitProtocol(config config.Config) {
	_, err := internal.CreateAllocator(config.CIDR)
	if err != nil {
		log.Fatal(err)
	}
	tun.CreateTunInterface(config)
}

func StartServer(config config.Config) {
	iface, err := tun.CreateTunInterface(config)
	if err != nil {
		log.Fatalf("failed to create tun device: %v", err)
	}
	allocator, err := internal.CreateAllocator(config.CIDR)
	if err != nil {
		log.Fatal(err)
	}

	initAPIRoutes(config, allocator)
	initWebSocket(config, iface)

	log.Printf("Starting server on: %v...", config.ServerAddr)
	log.Fatal(http.ListenAndServe(config.ServerAddr, nil))
}

// checkPermission checks the permission of the request
func checkPermission(w http.ResponseWriter, req *http.Request, config config.Config) bool {
	if config.Key == "" {
		return true
	}
	key := req.Header.Get("key")
	if key != config.Key {
		response := ErrorResponse{
			Message: "not permitted",
		}
		sendJsonResponse(w, http.StatusForbidden, response)
		return false
	}
	return true
}
