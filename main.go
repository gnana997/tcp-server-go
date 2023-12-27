package main

import (
	"fmt"
	"net"
)

type Message struct {
	from    string
	payload []byte
}

type Server struct {
	listenAddr string
	ln         net.Listener
	// channel with struct as empty struct takes no space
	quitch chan struct{}
	// channel with byte type as it can accept any data type
	msgCh chan Message
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		// adding the buffer of 10 to the channel as it will block the the go routines
		// until the msg gets processed from the channel
		msgCh: make(chan Message, 10),
	}
}

func (s *Server) Start() error {
	// creating a listener with tcp protocol on the listen addr provided to the server
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}

	// close the listener once everything is completed
	defer ln.Close()

	s.ln = ln

	// spin up a go routine to run the accept loop function
	// to accept the new connections to our tcp server
	go s.acceptLoop()

	//wait for the quitch channel
	<-s.quitch

	// close the msg channel to let the consumers of msg channel know that the channel is closed
	close(s.msgCh)

	// if the quitch channel is completed then return nil and listener will get closed
	return nil
}

// function to accept all the connections to our server
func (s *Server) acceptLoop() {
	for {
		// Accept the incoming connection
		conn, err := s.ln.Accept()
		if err != nil {
			// If error occurs while accepting connection print the error
			fmt.Println("accept error: ", err)
			// to accept the other incoming connections
			continue
		}

		// just to print the new connection remote address
		fmt.Println("Accepted a new connection from: ", conn.RemoteAddr())

		// running readLoop function on each connection as a go routine
		// to support multiple connection at a time
		go s.readLoop(conn)
	}
}

// function to read from the connections
func (s *Server) readLoop(conn net.Conn) {

	// Close the connection once the messages sent by the connection are read
	defer conn.Close()

	// making a buffer to read the files from the connection
	buf := make([]byte, 2048)

	for {
		// read the message sent by the connecion in to the buffer
		n, err := conn.Read(buf)
		if err != nil {
			// error occured while reading the message
			fmt.Println("read error: ", err)
			continue
		}

		// send the message and from details to the channel
		s.msgCh <- Message{
			from:    conn.RemoteAddr().String(),
			payload: buf[:n],
		}
	}
}

func main() {
	server := NewServer(":3000")

	// unblocking function which continously consumes the messages
	// sent to the server from the connections
	go func() {
		for msg := range server.msgCh {
			fmt.Printf("recieved message from connection (%s): %s\n", msg.from, string(msg.payload))
		}
	}()

	server.Start()
}
