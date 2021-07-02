package buf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const SELF_SP_ID = 666

const (
	ErrInvalidParam = "Parameter is invalid"
)

type OpError struct {
	err error
	op  string
}

// 定长字符串(右补0)
// 1.若左补0，则补ASCII表示的0
// 2.若右补0，则补二进制0
func OctetString(s string, fixedLength int) string {
	length := len(s)
	if length == fixedLength {
		return s
	}

	if length > fixedLength {
		return s[length-fixedLength:]
	}

	return strings.Join([]string{s, string(make([]byte, fixedLength-length))}, "")
}

func (e *OpError) Error() string {
	if e.err == nil {
		return "<nil>"
	}

	return e.op + "error:" + e.err.Error()
}

type BufWriter struct {
	buf *bytes.Buffer
	err *OpError
}

func NewBufWriter(length uint32) *BufWriter {
	buf := make([]byte, 0, length)
	return &BufWriter{
		buf: bytes.NewBuffer(buf),
	}
}

func (t *BufWriter) Bytes() ([]byte, error) {
	if t.err != nil {
		return nil, t.err
	}

	len := t.buf.Len()

	return (t.buf.Bytes())[:len], nil
}

func (t *BufWriter) WriteByte(b byte) {
	if t.err != nil {
		return
	}

	err := t.buf.WriteByte(b)
	if err != nil {
		t.err = &OpError{
			err: err,
			op:  fmt.Sprintf("write %x", b),
		}
	}
}

func (t *BufWriter) WriteFixedSizeString(s string, size int) {
	if t.err != nil {
		return
	}

	sLen := len(s)
	if sLen > size {
		t.err = &OpError{
			err: fmt.Errorf(ErrInvalidParam + ":" + s),
			op:  fmt.Sprintf("write %s", s),
		}
		return
	}

	n, err := t.buf.WriteString(strings.Join([]string{s, string(make([]byte, size-sLen))}, ""))
	if err != nil {
		t.err = &OpError{
			err: err,
			op:  fmt.Sprintf("write %s", s),
		}
		return
	}

	if n != size {
		t.err = &OpError{
			err: fmt.Errorf("write %d bytes, not equal to %d we expected", n, size),
			op:  fmt.Sprintf("write %s", s),
		}
	}
}

func (t *BufWriter) WriteString(s string) {
	if t.err != nil {
		return
	}

	sLen := len(s)

	n, err := t.buf.WriteString(s)
	if err != nil {
		t.err = &OpError{
			err: err,
			op:  fmt.Sprintf("write %s", s),
		}
		return
	}

	if n != sLen {
		t.err = &OpError{
			err: fmt.Errorf("write %d bytes, not equal to %d we expected", n, sLen),
			op:  fmt.Sprintf("write %s", s),
		}
	}
}

func (t *BufWriter) WriteInt(data interface{}, size int, order binary.ByteOrder) {
	if t.err != nil {
		return
	}

	err := binary.Write(t.buf, order, data)
	if err != nil {
		t.err = &OpError{
			err: err,
			op:  fmt.Sprintf("write %v", data),
		}
	}
}

// func GetMsgId(svrSpId int, seqId int) {
// 	appender := NewBufWriter()

// 	now := time.Now()
// 	nowMonth, _ := strconv.Atoi(fmt.Sprintf("%d", now.Month()))
// 	nowDay := now.Day()
// 	nowHourt := now.Hour()
// 	nowMin := now.Minute()
// 	nowSec := now.Second()

// }

func GetNow() (string, int) {
	s := time.Now().Format("0102150405")
	i, _ := strconv.Atoi(s)
	return s, i
}
