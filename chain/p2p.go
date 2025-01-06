package chain

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// Stream manages the P2P streaming connection
type Stream struct {
	Host              string
	Port              int
	Peers             []string
	listener          net.Listener
	activeConnections sync.Map
}

// NewStream creates a new P2P stream
func NewStream(host string, port int) (*Stream, error) {
	return &Stream{
		Host:              host,
		Port:              port,
		Peers:             []string{},
		activeConnections: sync.Map{},
	}, nil
}

// Start starts the P2P server
func (s *Stream) Start(data []byte) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	go s.acceptConnections(data)

	return nil
}

// acceptConnections accepts new connections
func (s *Stream) acceptConnections(data []byte) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			return
		}

		remoteAddr := conn.RemoteAddr().String()
		s.activeConnections.Store(remoteAddr, true)

		go s.handleConnection(conn, data)
	}
}

// handleConnection handles a single connection
func (s *Stream) handleConnection(conn net.Conn, data []byte) {
	defer conn.Close()
	defer func() {
		remoteAddr := conn.RemoteAddr().String()
		s.activeConnections.Delete(remoteAddr)
	}()

	writer := bufio.NewWriter(conn)
	_, err := writer.Write(data)
	if err != nil {
		log.Printf("failed to write data to connection: %v", err)
		return
	}
	writer.Flush()

	log.Println("data streamed to:", conn.RemoteAddr())

	io.Copy(io.Discard, conn) // keep connection alive till client disconnects
}

// Connect connects to a peer
func (s *Stream) Connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	s.Peers = append(s.Peers, address)
	go func() {
		defer conn.Close()
		io.Copy(io.Discard, conn) // keep connection alive till disconnection
	}()
	return nil
}

// Stop stops the P2P server
func (s *Stream) Stop() error {
	err := s.listener.Close()
	if err != nil {
		return fmt.Errorf("failed to close listener: %w", err)
	}
	s.activeConnections.Range(func(key, value any) bool {
		if remoteAddr, ok := key.(string); ok {
			fmt.Println("Disconnecting from ", remoteAddr)
			s.activeConnections.Delete(remoteAddr)
		}
		return true
	})
	return nil
}
