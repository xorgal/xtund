// File: server/packet.go
package server

import (
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/snappy"
	"github.com/net-byte/water"
	"github.com/xorgal/xtun-core/pkg/cache"
	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtun-core/pkg/counter"
	"github.com/xorgal/xtun-core/pkg/netutil"
	"github.com/xorgal/xtund/internal"
)

// toClient sends data to client
func toClient(config config.Config, iface *water.Interface) {
	packet := make([]byte, config.BufferSize)
	for {
		n, err := iface.Read(packet)
		if err != nil {
			internal.PrintErr("iface.Read(packet):", err)
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
			internal.PrintErr("wsutil.ReadClientData(wsconn)", err)
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
