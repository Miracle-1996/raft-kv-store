// Package http provides the HTTP server for accessing the distributed key-value store.
// It also provides the endpoint for other nodes to join an existing cluster.
package http

import (
	"bytes"
	"fmt"
	"github.com/RAFT-KV-STORE/raftpb"
	"github.com/RAFT-KV-STORE/store"
	"github.com/gogo/protobuf/proto"
	"log"
	"net"
	"net/http"
	"strings"
)

//// TransactionOps ...
//type TransactionOps struct {
//	Commands []store.Ops `json:"Commands"`
//}


// Service provides HTTP service.
type Service struct {
	addr string
	ln   net.Listener
	store *store.Store
}

// New returns an uninitialized HTTP service.
func NewService(addr string, store *store.Store) *Service {
	return &Service{
		addr:  addr,
		store: store,
	}
}

// Start starts the service.
func (s *Service) Start(joinHttpAddress string) {
	server := http.Server{
		Handler: s,
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}
	s.ln = ln

	http.Handle("/", s)

	go func() {
		err := server.Serve(s.ln)
		if err != nil {
			log.Fatalf("HTTP serve error: %s", err)
		}
	}()

	if joinHttpAddress != "" {
		msg := &raftpb.JoinMsg{RaftAddress: s.store.RaftAddress, ID: s.store.ID}
		b, err := proto.Marshal(msg)
		if err != nil {
			log.Fatalf("error when marshaling %+v", msg)
		}
		resp, err := http.Post(fmt.Sprintf("http://%s/join", joinHttpAddress), "application/protobuf", bytes.NewBuffer(b))
		if err != nil {
			log.Fatalf("failed to join %s: %s", joinHttpAddress, err)
		}
		defer resp.Body.Close()
	}
}

// Close closes the service.
func (s *Service) Close() {
	s.ln.Close()
	return
}

// ServeHTTP allows Service to serve HTTP requests.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/key") {
		s.handleKeyRequest(w, r)
	} else if r.URL.Path == "/join" {
		s.handleJoin(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/leader") {
		s.handleLeader(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/transaction") {
		s.handleTransaction(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}