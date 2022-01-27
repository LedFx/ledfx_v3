package rtsp

import (
	"errors"
	"fmt"
	"io"
	log "ledfx/logger"
	"net"
	"tailscale.com/net/interfaces"
)

// RequestHandler callback function that gets invoked when a request is received
type RequestHandler func(req *Request, resp *Response, localAddr string, remoteAddr string)

// Server is for handling RTSP control requests
type Server struct {
	port     int
	handlers map[Method]RequestHandler
	done     chan bool
	reqChan  chan *Request
	ip       string
}

// NewServer instantiates a new RtspServer
func NewServer(port int) *Server {
	return &Server{
		port:     port,
		done:     make(chan bool),
		handlers: make(map[Method]RequestHandler),
		reqChan:  make(chan *Request),
	}
}

// AddHandler registers a handler for a given RTSP method
func (r *Server) AddHandler(m Method, rh RequestHandler) {
	r.handlers[m] = rh
}

// Stop stops the RTSP server
func (r *Server) Stop() {
	log.Logger.WithField("category", "RTSP Server").Println("Stopping RTSP server")
	r.done <- true
}

// Start creates listening socket for the RTSP connection
func (r *Server) Start(verbose bool) {
	// Get the default outbound interface address
	_, myIP, ok := interfaces.LikelyHomeRouterIP()
	if !ok {
		log.Logger.WithField("category", "RTSP Server").Errorf("Error getting local outbound IP address: ok=%v", ok)
		return
	}
	r.ip = myIP.String()
	log.Logger.WithField("category", "RTSP Server").Printf("Starting RTSP server on address: %s:%d", r.ip, r.port)

	tcpListen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", r.ip, r.port))
	if err != nil {
		log.Logger.WithField("category", "RTSP Server").Errorln("Error listening:", err.Error())
		return
	}

	defer tcpListen.Close()

	// Handle TCP connections.
	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := tcpListen.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Logger.WithField("category", "RTSP Server").Warnf("Error accepting: %v", err)
				}
				return
			}
			go r.read(conn, r.handlers, verbose)
		}
	}()

	<-r.done
}

func (r *Server) read(conn net.Conn, handlers map[Method]RequestHandler, verbose bool) {
	defer conn.Close()
	for {
		request, err := readRequest(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Logger.WithField("category", "RTSP Server").Println("Client closed connection")
			} else {
				log.Logger.WithField("category", "RTSP Server").Println("Error reading data: ", err.Error())
			}
			return
		}

		if verbose {
			log.Logger.WithField("category", "RTSP Server").Println("Received Request")
			log.Logger.WithField("category", "RTSP Server").Println(request.String())
		}

		handler, exists := handlers[request.Method]
		if !exists {
			log.Logger.WithField("category", "RTSP Server").Printf("Method: %s does not have a handler. Skipping", request.Method)
			continue
		}
		resp := NewResponse()
		// for now we just stick in the protocol (protocol/version) from the request
		resp.protocol = request.protocol
		// same with CSeq
		resp.Headers["CSeq"] = request.Headers["CSeq"]
		// invokes the client specified handler to build the response
		localAddr := conn.LocalAddr().(*net.TCPAddr).IP
		remoteAddr := conn.RemoteAddr().(*net.TCPAddr).IP
		handler(request, resp, localAddr.String(), remoteAddr.String())
		if verbose {
			log.Logger.WithField("category", "RTSP Server").Println("Outbound Response")
			log.Logger.WithField("category", "RTSP Server").Println(resp.String())
		}

		_, _ = writeResponse(conn, resp)
	}
}
