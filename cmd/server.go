package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	ws "github.com/adriancable/webtransport-go"
)

type IWSServer interface {
	Run() error
}

type WSServer struct {
	ctx    context.Context
	server *ws.Server
}

func New(ctx context.Context) *WSServer {
	s := WSServer{
		ctx: ctx,
		server: &ws.Server{
			ListenAddr:     ":4433",
			AllowedOrigins: []string{"googlechrome.github.io", "127.0.0.1:8000", "localhost:8000", "new-tab-page", ""},
			TLSCert:        ws.CertFile{Path: `/Users/matiaet98/fake_credentials/selfsigned.pem`},
			TLSKey:         ws.CertFile{Path: `/Users/matiaet98/fake_credentials/selfsigned.key`},
			QuicConfig: &ws.QuicConfig{
				KeepAlive:      true,
				MaxIdleTimeout: 30 * time.Second,
			},
		},
	}
	return &s
}

func (w *WSServer) RegisterHandler() {
	http.HandleFunc("/counter", func(rw http.ResponseWriter, r *http.Request) {
		session := r.Body.(*ws.Session)
		session.AcceptSession()
		// session.RejectSession(400)

		fmt.Println("Accepted incoming WebTransport session")
		w.handleWebTransportStreams(session)

		s, err := session.OpenStreamSync(session.Context())
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("Listening on server-initiated bidi stream %v\n", s.StreamID())

		sendMsg := []byte("bidi")
		fmt.Printf("Sending to server-initiated bidi stream %v: %s\n", s.StreamID(), sendMsg)
		s.Write(sendMsg)
		go func(s ws.Stream) {
			defer s.Close()
			for {
				buf := make([]byte, 1024)
				n, err := s.Read(buf)
				if err != nil {
					log.Printf("Error reading from server-initiated bidi stream %v: %v\n", s.StreamID(), err)
					break
				}
				fmt.Printf("Received from server-initiated bidi stream %v: %s\n", s.StreamID(), buf[:n])
			}
		}(s)

		sUni, err := session.OpenUniStreamSync(session.Context())
		if err != nil {
			log.Println(err)
		}

		sendMsg = []byte("uni")
		fmt.Printf("Sending to server-initiated uni stream %v: %s\n", s.StreamID(), sendMsg)
		sUni.Write(sendMsg)
	})

}

func (w *WSServer) handleWebTransportStreams(session *ws.Session) {
	// Handle incoming datagrams
	go func() {
		for {
			msg, err := session.ReceiveMessage(session.Context())
			if err != nil {
				fmt.Println("Session closed, ending datagram listener:", err)
				break
			}
			fmt.Printf("Received datagram: %s\n", msg)

			sendMsg := bytes.ToUpper(msg)
			fmt.Printf("Sending datagram: %s\n", sendMsg)
			session.SendMessage(sendMsg)
		}
	}()

	// Handle incoming unidirectional streams
	go func() {
		for {
			s, err := session.AcceptUniStream(session.Context())
			if err != nil {
				fmt.Println("Session closed, not accepting more uni streams:", err)
				break
			}
			fmt.Println("Accepting incoming uni stream:", s.StreamID())

			go func(s ws.ReceiveStream) {
				for {
					buf := make([]byte, 1024)
					n, err := s.Read(buf)
					if err != nil {
						log.Printf("Error reading from uni stream %v: %v\n", s.StreamID(), err)
						break
					}
					fmt.Printf("Received from uni stream: %s\n", buf[:n])
				}
			}(s)
		}
	}()

	// Handle incoming bidirectional streams
	go func() {
		for {
			s, err := session.AcceptStream()
			if err != nil {
				fmt.Println("Session closed, not accepting more bidi streams:", err)
				break
			}
			fmt.Println("Accepting incoming bidi stream:", s.StreamID())

			go func(s ws.Stream) {
				defer s.Close()
				for {
					buf := make([]byte, 1024)
					n, err := s.Read(buf)
					if err != nil {
						log.Printf("Error reading from bidi stream %v: %v\n", s.StreamID(), err)
						break
					}
					fmt.Printf("Received from bidi stream %v: %s\n", s.StreamID(), buf[:n])
					sendMsg := bytes.ToUpper(buf[:n])
					fmt.Printf("Sending to bidi stream %v: %s\n", s.StreamID(), sendMsg)
					s.Write(sendMsg)
					// session.CloseSession()
					// session.CloseWithError(1234, "error")
				}
			}(s)
		}
	}()
}

func (w *WSServer) Run() error {

	return nil
}
