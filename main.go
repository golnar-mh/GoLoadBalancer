package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type Node struct {
	mu               sync.Mutex
	activeRequests   int
	maxCPU           int
	memoryUsage      int
	maxMemory        int
	memoryPerRequest int
	peers            []*url.URL
	hostname         string
}

type BlockData struct {
	BlockNumber int    `json:"blockNumber"`
	Timestamp   int64  `json:"timestamp"`
	Hash        string `json:"hash"`
}

func (n *Node) log(format string, args ...interface{}) {
	prefix := fmt.Sprintf("[%s] ", n.hostname)
	log.Printf(prefix+format, args...)
}

func (n *Node) ethBlockHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	const maxHops = 5 // Prevent infinite proxy loops

	// Get or initialize request chain
	requestChain := r.Header.Get("X-Request-Chain")
	if requestChain == "" {
		requestChain = n.hostname
	} else {
		// Check if we've exceeded hop limit or node is already in chain
		hops := strings.Split(requestChain, " → ")
		if len(hops) >= maxHops || strings.Contains(requestChain, n.hostname) {
			n.log("❌ %s - Hop limit exceeded or loop detected", requestChain)
			http.Error(w, "Too many hops or loop detected", http.StatusTooManyRequests)
			return
		}
		requestChain += " → " + n.hostname
	}

	// Resource check
	n.mu.Lock()
	currentCPU := n.activeRequests
	currentMemory := n.memoryUsage
	requiredMemory := n.memoryPerRequest

	canHandle := currentCPU < n.maxCPU && (currentMemory+requiredMemory) <= n.maxMemory

	if canHandle {
		n.activeRequests++
		n.memoryUsage += requiredMemory
		n.mu.Unlock()

		defer func() {
			n.mu.Lock()
			n.activeRequests--
			n.memoryUsage -= requiredMemory
			n.mu.Unlock()
		}()

		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

		block := BlockData{
			BlockNumber: rand.Intn(1000000),
			Timestamp:   time.Now().Unix(),
			Hash:        fmt.Sprintf("0x%032x", rand.Uint64()),
		}

		response := map[string]interface{}{
			"node":    n.hostname,
			"chain":   requestChain,
			"block":   block,
			"message": "Handled directly",
		}

		n.log("✅ %s - Handled (CPU: %d/%d, Mem: %d/%d MB) Time: %v",
			requestChain, currentCPU+1, n.maxCPU,
			currentMemory+requiredMemory, n.maxMemory, time.Since(start))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Can't handle locally - try to proxy
	reason := ""
	if currentCPU >= n.maxCPU {
		reason = fmt.Sprintf("CPU overload (%d/%d)", currentCPU, n.maxCPU)
	} else {
		reason = fmt.Sprintf("Memory overload (%d+%d > %d)",
			currentMemory, requiredMemory, n.maxMemory)
	}
	n.mu.Unlock()

	if len(n.peers) == 0 {
		n.log("🚨 %s - No peers available! (%s)", requestChain, reason)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	peer := n.peers[rand.Intn(len(n.peers))]
	proxy := httputil.NewSingleHostReverseProxy(peer)
	proxy.Director = func(req *http.Request) {
		req.Header.Set("X-Request-Chain", requestChain)
		req.URL.Scheme = peer.Scheme
		req.URL.Host = peer.Host
		req.Host = peer.Host
	}

	n.log("🔀 %s - Proxying to %s (Reason: %s)", requestChain, peer.Host, reason)
	proxy.ServeHTTP(w, r)
}

func (n *Node) healthHandler(w http.ResponseWriter, r *http.Request) {
	n.mu.Lock()
	defer n.mu.Unlock()

	status := "ready"
	if n.activeRequests >= n.maxCPU || n.memoryUsage >= n.maxMemory {
		status = "busy"
	}

	response := map[string]interface{}{
		"hostname":       n.hostname,
		"activeRequests": n.activeRequests,
		"maxCPU":         n.maxCPU,
		"memoryUsage":    n.memoryUsage,
		"maxMemory":      n.maxMemory,
		"status":         status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	hostname, _ := os.Hostname()

	node := &Node{
		hostname:         hostname,
		maxCPU:           getEnvInt("MAX_CPU", 2),
		maxMemory:        getEnvInt("MAX_MEMORY", 200),
		memoryPerRequest: getEnvInt("MEMORY_PER_REQUEST", 50),
	}

	if peers := os.Getenv("PEER_NODES"); peers != "" {
		for _, p := range strings.Split(peers, ",") {
			u, err := url.Parse(p)
			if err != nil {
				log.Fatalf("Invalid peer URL: %v", err)
			}
			node.peers = append(node.peers, u)
		}
	}

	router := mux.NewRouter()
	router.HandleFunc("/eth/block", node.ethBlockHandler)
	router.HandleFunc("/health", node.healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	node.log("🚀 Starting node (CPU: %d/%d, Mem: %d/%d MB) on :%s with %d peers",
		node.activeRequests, node.maxCPU, node.memoryUsage,
		node.maxMemory, port, len(node.peers))
	log.Fatal(server.ListenAndServe())
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
