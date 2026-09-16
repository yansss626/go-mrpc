package mrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

// unary
type Client struct {
	conn          net.Conn
	closed        atomic.Bool
	requestID     uint64
	unaryPending  map[uint64]chan unaryResult
	streamPending map[uint64]*ClientStream
	mu            sync.Mutex
	writeMu       sync.Mutex
}
type unaryResult struct {
	response *Response
	err      error
}

// stream
type ClientStream struct {
	requestID uint64
	client    *Client
	recvCh    chan streamResult
	ctx       context.Context
}

type streamResult struct {
	response *Response
	err      error
}

func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	c := &Client{
		conn:          conn,
		unaryPending:  make(map[uint64]chan unaryResult),
		streamPending: make(map[uint64]*ClientStream),
	}
	go c.readLoop()
	return c, nil
}

func (c *Client) nextRequestID() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestID++
	return c.requestID
}

func (c *Client) readLoop() {

	for {
		header, payload, err := ReadFrame(c.conn)
		if err != nil {
			c.failAll(err)
			c.closed.Store(true)
			return
		}
		switch header.FrameType {

		case FrameUnaryResponse:
			c.handleUnaryResponse(header.RequestID, payload)
		case FrameStreamData:
			c.handleStreamData(header.RequestID, payload)
		case FrameStreamEnd:
			c.handleStreamEnd(header.RequestID, payload)
		default:
			c.failAll(fmt.Errorf("invalid frame type: %d", header.FrameType))
			c.closed.Store(true)
			return
		}

	}

}
func (c *Client) handleUnaryResponse(requestID uint64, payload []byte) {
	ch, exist := c.getUnaryPending(requestID)
	if !exist {
		return
	}
	resp, err := DecodeResponse(payload)
	ch <- unaryResult{
		response: resp,
		err:      err,
	}

}
func (c *Client) failAll(err error) {

	c.mu.Lock()
	unaryList := make([]chan unaryResult, 0, len(c.unaryPending))
	for id, ch := range c.unaryPending {
		unaryList = append(unaryList, ch)
		delete(c.unaryPending, id)
	}
	streamList := make([]*ClientStream, 0, len(c.streamPending))
	for id, stream := range c.streamPending {
		streamList = append(streamList, stream)
		delete(c.streamPending, id)
	}
	c.mu.Unlock()

	for _, ch := range unaryList {
		ch <- unaryResult{
			err: err,
		}
	}

	for _, stream := range streamList {
		stream.recvCh <- streamResult{
			err: err,
		}
	}

}
func (c *Client) CallUnary(ctx context.Context, service, method string, req any, resp any) error {

	requestID := c.nextRequestID()
	ch := make(chan unaryResult, 1)
	c.addUnaryPending(requestID, ch)

	defer c.delUnaryPending(requestID)

	mrpcRequest, err := EncodeRequest(service, method, req)
	if err != nil {
		return err
	}

	c.sendRequest(requestID, FrameUnaryRequest, mrpcRequest)
	if err != nil {
		return err
	}

	select {
	case unaryResult := <-ch:
		if unaryResult.err != nil {
			return unaryResult.err
		}
		mrpcResp := unaryResult.response
		if mrpcResp.Code != CodeOK {
			return fmt.Errorf("rpc error: code=%d message=%s", mrpcResp.Code, mrpcResp.Message)
		}

		if resp != nil {
			err := json.Unmarshal(mrpcResp.Response, resp)
			if err != nil {
				return err
			}
		}
	case <-ctx.Done():
		c.delUnaryPending(requestID)
		c.sendRequest(requestID, FrameCancel, nil)
		return ctx.Err()
	}

	return nil
}

func (c *Client) addUnaryPending(requestID uint64, ch chan unaryResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.unaryPending[requestID] = ch
}
func (c *Client) delUnaryPending(requestID uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.unaryPending, requestID)
}
func (c *Client) getUnaryPending(requestID uint64) (chan unaryResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ch, ok := c.unaryPending[requestID]
	return ch, ok
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	c.closed.Store(true)
	return c.conn.Close()
}

// stream
func (c *Client) addStreamPending(requestID uint64, stream *ClientStream) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.streamPending[requestID] = stream
}
func (c *Client) delStreamPending(requestID uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.streamPending, requestID)
}
func (c *Client) getStreamPending(requestID uint64) (*ClientStream, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	stream, ok := c.streamPending[requestID]
	return stream, ok
}
func (c *Client) handleStreamData(requestID uint64, payload []byte) {
	stream, exist := c.getStreamPending(requestID)
	if !exist {
		return
	}
	resp, err := DecodeResponse(payload)
	stream.recvCh <- streamResult{
		response: resp,
		err:      err,
	}
}
func (c *Client) handleStreamEnd(requestID uint64, payload []byte) {
	stream, exist := c.getStreamPending(requestID)
	if !exist {
		return
	}
	c.delStreamPending(requestID)
	if len(payload) == 0 {
		stream.recvCh <- streamResult{
			err: io.EOF,
		}
		return
	}
	resp, err := DecodeResponse(payload)
	if err != nil {
		stream.recvCh <- streamResult{
			response: resp,
			err:      err,
		}
		return
	}
	stream.recvCh <- streamResult{
		response: resp,
		err:      io.EOF,
	}

}

func (s *ClientStream) Recv(resp any) error {
	select {
	case streamResult := <-s.recvCh:
		if streamResult.err != nil {
			return streamResult.err
		}

		mrpcResp := streamResult.response
		if mrpcResp.Code != CodeOK {
			return fmt.Errorf("rpc error: code=%d message=%s", mrpcResp.Code, mrpcResp.Message)
		}

		if resp != nil {
			err := json.Unmarshal(mrpcResp.Response, resp)
			if err != nil {
				return err
			}
		}
	case <-s.ctx.Done():
		s.client.sendRequest(s.requestID, FrameCancel, nil)
		return s.ctx.Err()
	}
	return nil
}

func (c *Client) NewClientStream(ctx context.Context, service string, method string, req any) (*ClientStream, error) {

	mrpcRequest, err := EncodeRequest(service, method, req)
	if err != nil {
		return nil, err
	}

	requestID := c.nextRequestID()
	ch := make(chan streamResult, 32)
	stream := &ClientStream{
		requestID: requestID,
		client:    c,
		recvCh:    ch,
		ctx:       ctx,
	}
	c.addStreamPending(requestID, stream)

	err = c.sendRequest(requestID, FrameStreamOpen, mrpcRequest)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (c *Client) sendRequest(requestID uint64, framType FrameType, request []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	err := WriteFrame(c.conn, requestID, framType, request)
	return err
}
