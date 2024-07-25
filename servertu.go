package mbserver

import (
	"io"

	"github.com/hootrhino/goserial"
)

// ListenRTU starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTU(&serial.Config{Address: "/dev/ttyUSB0"})
func (s *Server) ListenRTU(serialConfig *serial.Config) (err error) {
	port, err := serial.Open(serialConfig)
	if err != nil {
		s.Logger.Errorf("failed to open %s: %v\n", serialConfig.Address, err)
	}
	s.ports = append(s.ports, port)

	s.portsWG.Add(1)
	go func() {
		defer s.portsWG.Done()
		s.acceptSerialRequests(port)
	}()

	return err
}

func (s *Server) acceptSerialRequests(port serial.Port) {
	buffer := make([]byte, 512)
SkipFrameError:
	for {
		select {
		case <-s.Ctx.Done():
			return
		case <-s.portsCloseChan:
			return
		default:
		}
		bytesRead, err := port.Read(buffer)
		if err != nil {
			if err != io.EOF {
				s.Logger.Errorf("serial read error %v\n", err)
			}
			return
		}

		if bytesRead != 0 {
			packet := buffer[:bytesRead]
			frame, err := NewRTUFrame(packet)
			if err != nil {
				s.Logger.Errorf("bad serial frame error %v\n", err)
				continue SkipFrameError
			}
			s.Logger.Debug("Read Frame:", frame)
			request := &Request{port, frame}
			s.requestChan <- request
		}
	}
}
