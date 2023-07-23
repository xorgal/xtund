// Todo: move all shared types to xtun-core mod

package internal

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/snappy"
	"github.com/net-byte/water"
	"github.com/xorgal/xtun-core/pkg/cache"
	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtun-core/pkg/counter"
	"github.com/xorgal/xtun-core/pkg/netutil"
	"github.com/xorgal/xtun-core/pkg/tun"
)

type DefaultResponse struct {
	Timestamp int64 `json:"timestamp"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type RegisterDeviceRequest struct {
	DeviceId string `json:"id"`
}

type RegisterDeviceResponse struct {
	Server string `json:"server"`
	Client string `json:"client"`
}

func InitProtocol(config config.Config) {
	_, err := CreateAllocator(config.CIDR)
	if err != nil {
		log.Fatal(err)
	}
	tun.CreateTunInterface(config)
}

func StartServer(config config.Config) {
	iface := tun.CreateTunInterface(config)
	allocator, err := CreateAllocator(config.CIDR)
	if err != nil {
		log.Fatal(err)
	}
	// server -> client
	go toClient(config, iface)
	// client -> server
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if !checkPermission(w, r, config) {
			return
		}
		wsconn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			PrintErr("ws.UpgradeHTTP(r, w)", err)
			return
		}
		toServer(config, wsconn, iface)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		currentTime := time.Now().Unix()
		response := DefaultResponse{
			Timestamp: currentTime,
		}
		sendJsonResponse(w, http.StatusOK, response)
	})
	http.HandleFunc("/allocator/register", func(w http.ResponseWriter, r *http.Request) {
		if !checkPermission(w, r, config) {
			return
		}
		// Decode the JSON request
		var request RegisterDeviceRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		client, serverIP, err := allocator.RegisterDevice(request.DeviceId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			PrintErr("allocator.RegisterDevice(request.DeviceId):", err)
		}
		response := RegisterDeviceResponse{
			Client: client,
			Server: serverIP,
		}
		sendJsonResponse(w, http.StatusOK, response)
	})
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		if !checkPermission(w, r, config) {
			return
		}
		io.WriteString(w, counter.PrintBytes(true))
	})
	log.Fatal(http.ListenAndServe(config.ServerAddr, nil))
	log.Printf("xtun server started on %v", config.ServerAddr)
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

// toClient sends data to client
func toClient(config config.Config, iface *water.Interface) {
	packet := make([]byte, config.BufferSize)
	for {
		n, err := iface.Read(packet)
		if err != nil {
			PrintErr("iface.Read(packet):", err)
			break
		}
		b := packet[:n]
		if key := netutil.GetDstKey(b); key != "" {
			if v, ok := cache.GetCache().Get(key); ok {
				if config.Compress {
					b = snappy.Encode(nil, b)
				}
				err := wsutil.WriteServerBinary(v.(net.Conn), b)
				if err != nil {
					cache.GetCache().Delete(key)
					continue
				}
				counter.IncrWrittenBytes(n)
			}
		}
	}
}

// toServer sends data to server
func toServer(config config.Config, wsconn net.Conn, iface *water.Interface) {
	defer wsconn.Close()
	for {
		b, op, err := wsutil.ReadClientData(wsconn)
		if err != nil {
			PrintErr("wsutil.ReadClientData(wsconn)", err)
			break
		}
		if op == ws.OpText {
			wsutil.WriteServerMessage(wsconn, op, b)
		} else if op == ws.OpBinary {
			if config.Compress {
				b, _ = snappy.Decode(nil, b)
			}
			if key := netutil.GetSrcKey(b); key != "" {
				cache.GetCache().Set(key, wsconn, 24*time.Hour)
				counter.IncrReadBytes(len(b))
				iface.Write(b)
			}
		}
	}
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
