// Package mbserver implements a Modbus server (slave).
package mbserver

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/hootrhino/goserial"
	logrus "github.com/sirupsen/logrus"
)

// Server is a Modbus slave with allocated memory for discrete inputs, coils, etc.
type Server struct {
	Ctx context.Context
	// Debug enables more verbose messaging.
	Logger           *logrus.Logger
	Debug            bool
	listeners        []net.Listener
	ports            []serial.Port
	portsWG          sync.WaitGroup
	portsCloseChan   chan struct{}
	requestChan      chan *Request
	function         [256](func(*Server, Framer) ([]byte, *Exception))
	callbacks        [](func(*Server, Framer))
	DiscreteInputs   []byte
	Coils            []byte
	HoldingRegisters []uint16
	InputRegisters   []uint16
}

// Request contains the connection and Modbus frame.
type Request struct {
	conn  io.ReadWriteCloser
	frame Framer
}

// NewServer creates a new Modbus server (slave).
func NewServerWithContext(Ctx context.Context) *Server {
	Server := NewServer()
	Server.Ctx = Ctx
	return Server
}
func NewServer() *Server {
	s := &Server{}

	// Allocate Modbus memory maps.
	s.DiscreteInputs = make([]byte, 65536)
	s.Coils = make([]byte, 65536)
	s.HoldingRegisters = make([]uint16, 65536)
	s.InputRegisters = make([]uint16, 65536)
	s.callbacks = []func(*Server, Framer){}
	s.function = [256]func(*Server, Framer) ([]byte, *Exception){}
	// Add default functions.
	s.function[1] = ReadCoils
	s.function[2] = ReadDiscreteInputs
	s.function[3] = ReadHoldingRegisters
	s.function[4] = ReadInputRegisters
	s.function[5] = WriteSingleCoil
	s.function[6] = WriteHoldingRegister
	s.function[15] = WriteMultipleCoils
	s.function[16] = WriteHoldingRegisters

	s.requestChan = make(chan *Request)
	s.portsCloseChan = make(chan struct{})
	s.ports = make([]serial.Port, 0)
	s.listeners = make([]net.Listener, 0)
	go s.handler()

	return s
}
func (s *Server) SetLogger(Logger *logrus.Logger) {
	s.Logger = Logger
}

// RegisterFunctionHandler override the default behavior for a given Modbus function.
func (s *Server) RegisterFunctionHandler(funcCode uint8, function func(*Server, Framer) ([]byte, *Exception)) {
	s.function[funcCode] = function
}

// OnRequest
func (s *Server) SetOnRequest(Func func(*Server, Framer)) {
	s.callbacks = append(s.callbacks, Func)
}

func (s *Server) handle(request *Request) Framer {
	var exception *Exception
	var data []byte

	response := request.frame.Copy()

	function := request.frame.GetFunction()
	if s.function[function] != nil {
		data, exception = s.function[function](s, request.frame)
		response.SetData(data)
	} else {
		exception = &IllegalFunction
	}
	if exception != &Success {
		response.SetException(exception)
	}
	for _, cb := range s.callbacks {
		cb(s, request.frame)
	}
	return response
}

// All requests are handled synchronously to prevent modbus memory corruption.
func (s *Server) handler() {
	for {
		request := <-s.requestChan
		response := s.handle(request)
		request.conn.Write(response.Bytes())
	}
}

// Close stops listening to TCP/IP ports and closes serial ports.
func (s *Server) Close() {
	for _, listen := range s.listeners {
		listen.Close()
	}
	select {
	case _, ok := <-s.portsCloseChan:
		if !ok {
			close(s.portsCloseChan)
		}
	default:
	}
	s.portsWG.Wait()
	for _, port := range s.ports {
		port.Close()
	}
}
