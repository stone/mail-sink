// Very simple server that mimics a smtp-server, it is very forgiving
// about what you send to and tends to agree (OK) most of the time.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

var listenPort = flag.Int("p", 25, "listen port")
var listenIntf = flag.String("i", "localhost", "listen on interface")
var heloHostname = flag.String("H", "localhost", "hostname to greet with")
var logBody = flag.Bool("v", false, "log the mail body")

type SinkServer struct {
	service  string
	listener *net.TCPListener
	Stats    *SinkStats
}

type SinkStats struct {
	AcceptedConnetions int
}

type SinkClient struct {
	conn     *textproto.Conn
	dataSent bool
}

func NewSinkClient(conn *textproto.Conn) *SinkClient {
	return &SinkClient{conn: conn, dataSent: false}
}

func NewSinkServer(listenIntf string, listenPort int) (*SinkServer, error) {
	s := new(SinkServer)
	stats := new(SinkStats)
	service := fmt.Sprintf("%s:%v", listenIntf, listenPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		log.Fatalln("Could not start server: ", err)
		return nil, err
	}
	s.listener, err = net.ListenTCP("tcp", tcpAddr)
	s.Stats = stats
	return s, nil
}

func (s *SinkServer) ListenAndServe() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Fatalln("error accepting client: ", err)
			continue
		}
		s.Stats.AcceptedConnetions += 1
		go s.HandleClient(conn)
	}
	return nil
}

func (s *SinkServer) HandleClient(conn net.Conn) {
	sc := NewSinkClient(textproto.NewConn(conn))
	remoteClient := conn.RemoteAddr().String()

	defer func() {
		log.Printf("Closing connection to %s\r\n", remoteClient)
		sc.conn.Close()
	}()

	// Start off greeting our client
	greet := fmt.Sprintf("%s SMTP mail-sink", *heloHostname)
	respond(sc.conn, 220, greet)

	for {
		line, err := sc.conn.ReadLine()
		if err != nil {
			log.Println("error ReadLine:", remoteClient, err.Error())
			return
		}

		line = strings.ToLower(strings.TrimSpace(line))
		if sc.dataSent == false {
			log.Println(remoteClient + ": " + strconv.Quote(line))
		} else if *logBody {
			log.Println(remoteClient + ": " + strconv.Quote(line))
		}

		// We handle quit as a special case here..
		if line == "quit" {
			respond(sc.conn, 221, "Bye")
			return
		}

		code, msg := sc.handleQuery(line)
		if code != 0 {
			respond(sc.conn, code, msg)
		}
	}

}

func (s *SinkClient) handleQuery(line string) (code int, reply string) {

	if strings.HasPrefix(line, "data") {
		s.dataSent = true
		return 354, "End data with <CR><LF>.<CR><LF>"
	}

	if s.dataSent == true {
		if strings.HasPrefix(line, ".") {
			s.dataSent = false
			return 250, "Ok: queued as 31337"
		} else {
			return 0, ""
		}
	}

	return 250, "Ok"
}

func respond(conn *textproto.Conn, code int, msg string) error {
	return conn.PrintfLine("%d %s", code, msg)
}

func main() {
	flag.Parse()
	log.Printf("Starting mail-sink on %s:%d", *listenIntf, *listenPort)
	sink, err := NewSinkServer(*listenIntf, *listenPort)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Printf("Stats: %d connetions", sink.Stats.AcceptedConnetions)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	sink.ListenAndServe()
}
