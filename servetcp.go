package mbserver

import (
	"crypto/tls"
	"io"
	"net"
	"strings"
)

func (s *Server) accept(listen net.Listener) error {
	for {
		conn, err := listen.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			s.Logger.Errorf("Unable to accept connections: %#v", err)
			return err
		}
		s.Logger.Debug("Tcp connected:", conn.RemoteAddr())
		go func(conn net.Conn) {
			defer conn.Close()
			defer func() {
				s.Logger.Debug("Tcp disconnected:", conn.RemoteAddr())
			}()
			packet := make([]byte, 512)
			for {
				select {
				case <-s.Ctx.Done():
					return
				default:
				}
				bytesRead, err := conn.Read(packet)
				if err != nil {
					if err != io.EOF {
						s.Logger.Errorf("read error %v", err)
					}
					return
				}
				// Set the length of the packet to the number of read bytes.
				packet = packet[:bytesRead]

				frame, err := NewTCPFrame(packet)
				if err != nil {
					s.Logger.Errorf("bad packet error %v", err)
					return
				}
				s.Logger.Debug("Read Frame:", frame)
				request := &Request{conn, frame}
				s.requestChan <- request
			}
		}(conn)
	}
}

func (s *Server) ListenWithListener(listen net.Listener) {
	s.listeners = append(s.listeners, listen)
	go s.accept(listen)
}

// ListenTCP starts the Modbus server listening on "address:port".
func (s *Server) ListenTCP(addressPort string) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		s.Logger.Errorf("Failed to Listen: %v", err)
		return err
	}
	s.ListenWithListener(listen)
	return err
}

// ListenTLS starts the Modbus server listening on "address:port".
func (s *Server) ListenTLS(addressPort string, config *tls.Config) (err error) {
	listen, err := tls.Listen("tcp", addressPort, config)
	if err != nil {
		s.Logger.Errorf("Failed to Listen on TLS: %v", err)
		return err
	}
	s.ListenWithListener(listen)
	return err
}
