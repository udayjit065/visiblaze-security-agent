package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const dataDir = "data"

func ensureDataDir() error {
	return os.MkdirAll(dataDir, 0755)
}

func ingestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// try to extract host_id from body
	t := map[string]any{}
	if err := json.Unmarshal(body, &t); err == nil {
		// attempt nested access: host.host_id
		hostID := "unknown"
		if host, ok := t["host"].(map[string]any); ok {
			if hid, ok := host["host_id"].(string); ok && hid != "" {
				hostID = hid
			}
		}
		file := filepath.Join(dataDir, hostID+".json")
		if err := os.WriteFile(file, body, 0644); err != nil {
			log.Printf("failed to write payload: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","host_id":"` + hostID + `"}`))
		return
	}
	// fallback: write raw
	file := filepath.Join(dataDir, "raw_"+strings.ReplaceAll(filepath.Base(r.URL.Path), "/", "_")+".json")
	_ = os.WriteFile(file, body, 0644)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// withCORS is a small wrapper that sets CORS headers and handles preflight requests.
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
		// default to JSON responses
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func listHostsHandler(w http.ResponseWriter, r *http.Request) {
	files, _ := os.ReadDir(dataDir)
	hosts := []map[string]any{}
	for _, fi := range files {
		if fi.IsDir() || !strings.HasSuffix(fi.Name(), ".json") {
			continue
		}
		b, _ := os.ReadFile(filepath.Join(dataDir, fi.Name()))
		var payload map[string]any
		if err := json.Unmarshal(b, &payload); err != nil {
			continue
		}
		if host, ok := payload["host"].(map[string]any); ok {
			hosts = append(hosts, host)
		}
	}
	resp := map[string]any{"hosts": hosts}
	enc := json.NewEncoder(w)
	enc.Encode(resp)
}

func hostDetailHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hostID := parts[1]
	file := filepath.Join(dataDir, hostID+".json")
	b, err := os.ReadFile(file)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
		return
	}
	w.Write(b)
}

func appsHandler(w http.ResponseWriter, r *http.Request) {
	// aggregate packages from stored files
	files, _ := os.ReadDir(dataDir)
	apps := map[string]map[string]any{}
	for _, fi := range files {
		if fi.IsDir() || !strings.HasSuffix(fi.Name(), ".json") {
			continue
		}
		b, _ := os.ReadFile(filepath.Join(dataDir, fi.Name()))
		var payload map[string]any
		if err := json.Unmarshal(b, &payload); err != nil {
			continue
		}
		if packages, ok := payload["packages"].([]any); ok {
			for _, p := range packages {
				if pm, ok := p.(map[string]any); ok {
					name := ""
					if n, ok := pm["name"].(string); ok {
						name = n
					}
					apps[name] = pm
				}
			}
		}
	}
	list := []map[string]any{}
	for _, v := range apps {
		list = append(list, v)
	}
	json.NewEncoder(w).Encode(map[string]any{"apps": list})
}

func cisResultsHandler(w http.ResponseWriter, r *http.Request) {
	files, _ := os.ReadDir(dataDir)
	results := []map[string]any{}
	for _, fi := range files {
		if fi.IsDir() || !strings.HasSuffix(fi.Name(), ".json") {
			continue
		}
		b, _ := os.ReadFile(filepath.Join(dataDir, fi.Name()))
		var payload map[string]any
		if err := json.Unmarshal(b, &payload); err != nil {
			continue
		}
		if cis, ok := payload["cis_results"].([]any); ok {
			for _, c := range cis {
				if cm, ok := c.(map[string]any); ok {
					results = append(results, cm)
				}
			}
		}
	}
	json.NewEncoder(w).Encode(map[string]any{"cis_results": results})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	if err := ensureDataDir(); err != nil {
		log.Fatalf("failed to create data dir: %v", err)
	}

	http.HandleFunc("/ingest", withCORS(ingestHandler))
	http.HandleFunc("/hosts", withCORS(listHostsHandler))
	http.HandleFunc("/hosts/", withCORS(hostDetailHandler))
	http.HandleFunc("/apps", withCORS(appsHandler))
	http.HandleFunc("/cis-results", withCORS(cisResultsHandler))
	http.HandleFunc("/health", withCORS(healthHandler))

	addr := ":3001"
	log.Printf("mock server listening %s (data dir %s)", addr, dataDir)
	log.Fatal(http.ListenAndServe(addr, nil))
}
