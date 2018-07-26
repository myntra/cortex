package store

import (
	"errors"
	"io"
	"log"
	"net"
	"time"

	"github.com/hashicorp/raft"
)

var (
	errNotAdvertisable = errors.New("local bind address is not advertisable")
	errNotTCP          = errors.New("local address is not a TCP address")
)

// TCPStreamLayer implements StreamLayer interface for plain TCP.
type TCPStreamLayer struct {
	advertise net.Addr
	listener  *net.TCPListener
}

// NewTCPTransport returns a NetworkTransport that is built on top of
// a TCP streaming transport layer.
func NewTCPTransport(
	bindListener net.Listener,
	advertise net.Addr,
	maxPool int,
	timeout time.Duration,
	logOutput io.Writer,
) (*raft.NetworkTransport, error) {
	return newTCPTransport(bindListener, advertise, func(stream raft.StreamLayer) *raft.NetworkTransport {
		return raft.NewNetworkTransport(stream, maxPool, timeout, logOutput)
	})
}

// NewTCPTransportWithLogger returns a NetworkTransport that is built on top of
// a TCP streaming transport layer, with log output going to the supplied Logger
func NewTCPTransportWithLogger(
	bindListener net.Listener,
	advertise net.Addr,
	maxPool int,
	timeout time.Duration,
	logger *log.Logger,
) (*raft.NetworkTransport, error) {
	return newTCPTransport(bindListener, advertise, func(stream raft.StreamLayer) *raft.NetworkTransport {
		return raft.NewNetworkTransportWithLogger(stream, maxPool, timeout, logger)
	})
}

// NewTCPTransportWithConfig returns a NetworkTransport that is built on top of
// a TCP streaming transport layer, using the given config struct.
func NewTCPTransportWithConfig(
	bindListener net.Listener,
	advertise net.Addr,
	config *raft.NetworkTransportConfig,
) (*raft.NetworkTransport, error) {
	return newTCPTransport(bindListener, advertise, func(stream raft.StreamLayer) *raft.NetworkTransport {
		config.Stream = stream
		return raft.NewNetworkTransportWithConfig(config)
	})
}

func newTCPTransport(list net.Listener,
	advertise net.Addr,
	transportCreator func(stream raft.StreamLayer) *raft.NetworkTransport) (*raft.NetworkTransport, error) {

	// Create stream
	stream := &TCPStreamLayer{
		advertise: advertise,
		listener:  list.(*net.TCPListener),
	}

	// Verify that we have a usable advertise address
	addr, ok := stream.Addr().(*net.TCPAddr)
	if !ok {
		list.Close()
		return nil, errNotTCP
	}
	if addr.IP.IsUnspecified() {
		list.Close()
		return nil, errNotAdvertisable
	}

	// Create the network transport
	trans := transportCreator(stream)
	return trans, nil
}

// Dial implements the StreamLayer interface.
func (t *TCPStreamLayer) Dial(address raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", string(address), timeout)
}

// Accept implements the net.Listener interface.
func (t *TCPStreamLayer) Accept() (c net.Conn, err error) {
	return t.listener.Accept()
}

// Close implements the net.Listener interface.
func (t *TCPStreamLayer) Close() (err error) {
	return t.listener.Close()
}

// Addr implements the net.Listener interface.
func (t *TCPStreamLayer) Addr() net.Addr {
	// Use an advertise addr if provided
	if t.advertise != nil {
		return t.advertise
	}
	return t.listener.Addr()
}
