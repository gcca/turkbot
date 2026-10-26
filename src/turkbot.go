package turkbot

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

const protocolVersion = "2025-06-18"

func fail(w http.ResponseWriter, id json.RawMessage, code int, message string, status int) {
	w.WriteHeader(status)

	value := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	}

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}

func reply(w http.ResponseWriter, id json.RawMessage, result any) {
	w.WriteHeader(http.StatusOK)

	value := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response %v", err)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.ContentLength == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fail(w, nil, -32700, "Unable to read request.", http.StatusBadRequest)
		return
	}

	var message map[string]json.RawMessage
	if err = json.Unmarshal(body, &message); err != nil || message == nil {
		fail(w, nil, -32700, "Parse error.", http.StatusBadRequest)
		return
	}

	var version, method string
	if json.Unmarshal(message["jsonrpc"], &version) != nil || version != "2.0" || json.Unmarshal(message["method"], &method) != nil || method == "" {
		fail(w, nil, -32600, "Invalid JSON-RPC envelope.", http.StatusBadRequest)
		return
	}

	id, hasId := message["id"]
	header := r.Header.Get("MCP-Protocol-Version")
	if (method != "initialize" || header != "") && header != protocolVersion {
		http.Error(w, "Use MCP-Protocol-Version: "+protocolVersion, http.StatusBadRequest)
		return
	}

	if !hasId {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	switch method {
	case "initialize":
		reply(w, id, map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities":    map[string]any{"tool": map[string]any{}},
			"serverInfo":      map[string]any{"name": "turkbot", "version": "0.1.0"},
		})
	case "ping":
		reply(w, id, map[string]any{})
	default:
		fail(w, id, -32601, "Method not found.", http.StatusOK)
	}
}

func Run(host, port string) error {
	log.Println("Turkbot started")

	parsed := net.ParseIP(host)
	if host != "localhost" && (parsed == nil || !parsed.IsLoopback()) {
		return os.ErrInvalid
	}

	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return os.ErrInvalid
	}

	address := net.JoinHostPort(host, port)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /mcp", Handler)

	log.Printf("Turkbot endpoint http://%s/mcp", address)
	return http.ListenAndServe(address, mux)
}
