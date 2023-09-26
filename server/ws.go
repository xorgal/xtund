// File: server/ws.go
//
// Todo: Current implementation is missing initial handshake and packets encryption.
//
// Suggested implementation of handshake:
//
//  1. Implement protocol states: `HANDSHAKE`, `DATA`, "TERMINATED"
//  2. Set protocol state to `HANDSHAKE`
//  3. Public key exchange
//  4. Shared secret generation and derive symmetric key
//  5. Test encrypted packet exchange
//  6. Set shared secret in cache
//  7. Set protocol state to `DATA`
//  8. Set protocol state to `TERMINATED` if encryption related error occurs, close connection
//  9. Cleanup on disconnection, remove keys, close connection, log events
//
// This is basic implementation and still requires testing.
package server

import (
	"net/http"

	"github.com/gobwas/ws"
	"github.com/net-byte/water"
	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtund/internal"
)

func initWebSocket(config config.Config, iface *water.Interface) {
	go toClient(config, iface)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if !checkPermission(w, r, config) {
			return
		}
		wsconn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			internal.PrintErr("ws.UpgradeHTTP(r, w)", err)
			return
		}

		// Todo: handshake first

		toServer(config, wsconn, iface)
	})
}
