// main.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Config holds server configuration
type Config struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxConnections  int
	ShutdownTimeout time.Duration
}

// Message represents the JSON structure for client communication
type Message struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
	Time    time.Time              `json:"time"`
	ID      string                 `json:"id"`
	Source  string                 `json:"source"`
}

// Server handles all client connections and message processing
type Server struct {
	config    Config
	listener  net.Listener
	connMutex sync.RWMutex
	conns     map[net.Conn]struct{}
	shutdown  chan struct{}
	logger    *log.Logger
	connSem   chan struct{} // Semaphore for connection limiting
}

// NewServer creates and initializes a new server instance
func NewServer(config Config) *Server {
	return &Server{
		config:   config,
		conns:    make(map[net.Conn]struct{}),
		shutdown: make(chan struct{}),
		logger:   log.New(os.Stdout, "[SERVER] ", log.LstdFlags|log.Lmicroseconds),
		connSem:  make(chan struct{}, config.MaxConnections),
	}
}

// Start begins listening for connections
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", ":"+s.config.Port)
	if err != nil {
		return err
	}
	s.listener = listener
	s.logger.Printf("Server started on port %s", s.config.Port)

	go s.acceptConnections()
	return nil
}

// acceptConnections handles incoming client connections
func (s *Server) acceptConnections() {
	for {
		select {
		case <-s.shutdown:
			return
		case s.connSem <- struct{}{}: // Acquire semaphore slot
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.shutdown:
					return
				default:
					s.logger.Printf("Error accepting connection: %v", err)
					<-s.connSem // Release semaphore slot on error
					continue
				}
			}

			go s.handleConnection(conn)
		}
	}
}

// prettyPrintJSON formats a map for logging
func prettyPrintJSON(data map[string]interface{}, indent string) string {
	var result string
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			result += fmt.Sprintf("%s%s:\n%s", indent, key, prettyPrintJSON(v, indent+"  "))
		case []interface{}:
			result += fmt.Sprintf("%s%s: [array of %d elements]\n", indent, key, len(v))
		default:
			result += fmt.Sprintf("%s%s: %v\n", indent, key, value)
		}
	}
	return result
}

// handleConnection processes individual client connections
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		<-s.connSem // Release semaphore slot
		s.removeConnection(conn)
	}()

	s.addConnection(conn)
	remoteAddr := conn.RemoteAddr().String()
	s.logger.Printf("New connection from: %s", remoteAddr)

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			if err.Error() != "EOF" {
				s.logger.Printf("Error decoding message from %s: %v", remoteAddr, err)
			} else {
				s.logger.Printf("Connection closed by client: %s", remoteAddr)
			}
			return
		}

		// Log received message details
		s.logger.Printf("\nReceived message from %s:\n"+
			"╔══════════════════════════════\n"+
			"║ ID: %s\n"+
			"║ Type: %s\n"+
			"║ Source: %s\n"+
			"║ Time: %s\n"+
			"║ Payload:\n%s"+
			"╚══════════════════════════════",
			remoteAddr,
			msg.ID,
			msg.Type,
			msg.Source,
			msg.Time.Format(time.RFC3339Nano),
			prettyPrintJSON(msg.Payload, "║   "))

		// Process message
		msg.Time = time.Now()

		// Send response
		if err := encoder.Encode(msg); err != nil {
			s.logger.Printf("Error sending response to %s: %v", remoteAddr, err)
			return
		}
	}
}

// addConnection registers a new client connection
func (s *Server) addConnection(conn net.Conn) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()
	s.conns[conn] = struct{}{}
}

// removeConnection removes a client connection from tracking
func (s *Server) removeConnection(conn net.Conn) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()
	delete(s.conns, conn)
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	close(s.shutdown)

	// Stop accepting new connections
	if err := s.listener.Close(); err != nil {
		s.logger.Printf("Error closing listener: %v", err)
	}

	// Close all existing connections
	s.connMutex.Lock()
	for conn := range s.conns {
		if err := conn.Close(); err != nil {
			s.logger.Printf("Error closing connection: %v", err)
		}
	}
	s.connMutex.Unlock()

	// Wait for context timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func main() {
	// Command line flags
	port := flag.String("port", "8080", "Server port")
	maxConns := flag.Int("max-connections", 1000000, "Maximum concurrent connections")
	flag.Parse()

	config := Config{
		Port:            *port,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		MaxConnections:  *maxConns,
		ShutdownTimeout: 30 * time.Second,
	}

	server := NewServer(config)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
