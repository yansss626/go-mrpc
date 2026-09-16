package mrpc

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

const (
	HeaderLength     = 17
	MaxPayloadLength = 4 * 1024 * 1024
)

type FrameType uint8

const (
	FrameUnaryRequest  FrameType = 1 // 单次请求
	FrameUnaryResponse FrameType = 2 // 单次响应
	//stream
	FrameStreamOpen FrameType = 3
	FrameStreamData FrameType = 4
	FrameStreamEnd  FrameType = 5
)

type Header struct {
	CRC32         uint32
	RequestID     uint64
	PayloadLength uint32
	FrameType     FrameType
}

func EncodeHeader(h Header) []byte {
	buf := make([]byte, HeaderLength)

	binary.BigEndian.PutUint32(buf[0:4], h.CRC32)
	binary.BigEndian.PutUint64(buf[4:12], h.RequestID)
	binary.BigEndian.PutUint32(buf[12:16], h.PayloadLength)
	buf[16] = uint8(h.FrameType)
	return buf
}

func DecodeHeader(data []byte) (Header, error) {
	if len(data) != HeaderLength {
		return Header{}, fmt.Errorf("invalid header length: %d, expect: %d", len(data), HeaderLength)
	}

	return Header{
		CRC32:         binary.BigEndian.Uint32(data[0:4]),
		RequestID:     binary.BigEndian.Uint64(data[4:12]),
		PayloadLength: binary.BigEndian.Uint32(data[12:16]),
		FrameType:     FrameType(data[16]),
	}, nil
}
func ReadFrame(r io.Reader) (Header, []byte, error) {

	headerBuf := make([]byte, HeaderLength)

	_, err := io.ReadFull(r, headerBuf)
	if err != nil {
		return Header{}, nil, err
	}

	header, err := DecodeHeader(headerBuf)
	if err != nil {
		return Header{}, nil, err
	}

	if header.PayloadLength > MaxPayloadLength {
		return Header{}, nil, fmt.Errorf("invalid payload length: %d, max: %d", header.PayloadLength, MaxPayloadLength)
	}

	payloadBuf := make([]byte, header.PayloadLength)
	_, err = io.ReadFull(r, payloadBuf)
	if err != nil {
		return Header{}, nil, err
	}

	// crc32 check
	if crc32.ChecksumIEEE(payloadBuf) != header.CRC32 {
		return Header{}, nil, fmt.Errorf("crc32 mismatch")
	}

	return header, payloadBuf, nil
}

func WriteFrame(w io.Writer, requestID uint64, frameType FrameType, payload []byte) error {
	if len(payload) > MaxPayloadLength {
		return fmt.Errorf(
			"payload too large: %d, max: %d",
			len(payload),
			MaxPayloadLength,
		)
	}
	mrpcHeader := Header{
		CRC32:         crc32.ChecksumIEEE(payload),
		RequestID:     requestID,
		PayloadLength: uint32(len(payload)),
		FrameType:     frameType,
	}
	header := EncodeHeader(mrpcHeader)

	err := writeFull(w, header)
	if err != nil {
		return err
	}
	return writeFull(w, payload)
}
func writeFull(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}

		if n == 0 {
			return io.ErrUnexpectedEOF
		}

		data = data[n:]
	}

	return nil
}
