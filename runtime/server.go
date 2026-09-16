package mrpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Server struct {
	registry *Registry
}

type StreamServer struct {
	pro       *process
	requestID uint64
}

func NewServer(registry *Registry) *Server {
	return &Server{
		registry: registry,
	}
}

type process struct {
	conn    net.Conn
	writeMu sync.Mutex
}

func (s *Server) handler(ctx context.Context, pro *process) error {

	defer pro.conn.Close()
	for {
		header, payload, err := ReadFrame(pro.conn)
		if err != nil {
			return err
		}
		switch header.FrameType {
		// unary
		case FrameUnaryRequest:
			go func() {
				s.handleUnary(ctx, pro, header.RequestID, payload)
			}()
		//stream
		case FrameStreamOpen:
			go func() {
				s.handleStreamServer(ctx, pro, header.RequestID, payload)
			}()
		default:
		}
	}
}

func (s *Server) handleUnary(ctx context.Context, pro *process, requestID uint64, payload []byte) {
	processCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	req, err := DecodeRequest(payload)
	if err != nil {
		err := s.sendErr(pro, CodeBadRequest, err.Error(), requestID, FrameUnaryResponse)
		if err != nil {
			log.Println(err)
		}
		return
	}

	funcHandler, ok := s.registry.LookupUnary(req.Service, req.Method)
	if !ok || funcHandler == nil {
		message := fmt.Sprintf("method not found: %s:%s()", req.Service, req.Method)
		err := s.sendErr(pro, CodeMethodNotFound, message, requestID, FrameUnaryResponse)
		if err != nil {
			log.Println(err)
		}
		return
	}

	result, err := funcHandler(processCtx, req.Request)
	if err != nil {
		err := s.sendErr(pro, CodeError, err.Error(), requestID, FrameUnaryResponse)
		if err != nil {
			log.Println(err)
		}
		return
	}

	resp, err := EncodeResponse(CodeOK, "", result)
	if err != nil {
		err := s.sendErr(pro, CodeError, err.Error(), requestID, FrameUnaryResponse)
		if err != nil {
			log.Println(err)
		}
		return
	}
	pro.writeMu.Lock()
	err = WriteFrame(pro.conn, requestID, FrameUnaryResponse, resp)
	pro.writeMu.Unlock()
	if err != nil {
		return
	}
}
func (s *Server) handleStreamServer(ctx context.Context, pro *process, requestID uint64, payload []byte) {
	processCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	req, err := DecodeRequest(payload)
	if err != nil {
		err := s.sendErr(pro, CodeBadRequest, err.Error(), requestID, FrameStreamEnd)
		if err != nil {
			log.Println(err)
		}
		return
	}
	funcHandler, ok := s.registry.LookupStream(req.Service, req.Method)
	if !ok || funcHandler == nil {
		message := fmt.Sprintf("method not found: %s:%s()", req.Service, req.Method)
		err := s.sendErr(pro, CodeMethodNotFound, message, requestID, FrameStreamEnd)
		if err != nil {
			log.Println(err)
		}
		return
	}
	stream := &StreamServer{
		pro:       pro,
		requestID: requestID,
	}
	err = funcHandler(processCtx, req.Request, stream)
	if err != nil {
		err := s.sendErr(pro, CodeError, err.Error(), requestID, FrameStreamEnd)
		if err != nil {
			log.Println(err)
		}
		return
	}
	stream.close()
}

func (s *Server) sendErr(pro *process, code int32, message string, requestID uint64, frameType FrameType) error {

	resp, err := EncodeResponse(code, message, nil)
	if err != nil {
		return err
	}
	pro.writeMu.Lock()
	defer pro.writeMu.Unlock()

	return WriteFrame(pro.conn, requestID, frameType, resp)
}

func (s *Server) Listen(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func() {
			err := s.handler(context.Background(), &process{
				conn: conn,
			})
			if err != nil && err != io.EOF {
				log.Println(err)
			}
		}()
	}
}

// stream send
func (stream *StreamServer) Send(resp any) error {

	rpcResp, err := EncodeResponse(CodeOK, "", resp)
	if err != nil {
		return err
	}
	stream.pro.writeMu.Lock()
	defer stream.pro.writeMu.Unlock()

	return WriteFrame(stream.pro.conn, stream.requestID, FrameStreamData, rpcResp)
}

func (stream *StreamServer) close() error {
	stream.pro.writeMu.Lock()
	defer stream.pro.writeMu.Unlock()

	return WriteFrame(stream.pro.conn, stream.requestID, FrameStreamEnd, nil)
}
