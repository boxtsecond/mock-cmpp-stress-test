package buf

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type BufReader struct {
	buf *bytes.Buffer
	err *OpError
}

func NewBufReader(data []byte) *BufReader {
	return &BufReader{
		buf: bytes.NewBuffer(data),
	}
}

func (t *BufReader) ReadByte() byte {
	if t.err != nil {
		return 0
	}

	b, err := t.buf.ReadByte()
	if err != nil {
		t.err = &OpError{
			err: err,
			op:  fmt.Sprintf("ReadByte"),
		}
	}
	return b
}

func (t *BufReader) ReadInt(data interface{}, order binary.ByteOrder) {
	if t.err != nil {
		return
	}

	err := binary.Read(t.buf, order, data)
	if err != nil {
		t.err = &OpError{
			err: err,
			op:  fmt.Sprintf("ReadInt"),
		}
	}
}

func (t *BufReader) ReadBytes(s []byte) {
	if t.err != nil {
		return
	}

	n, err := t.buf.Read(s)
	if err != nil {
		t.err = &OpError{
			err: err,
			op:  fmt.Sprintf("ReadBytes"),
		}
		return
	}

	if n != len(s) {
		t.err = &OpError{
			err: fmt.Errorf("Read %d bytes, not equal to %d we expected", n, len(s)),
			op:  fmt.Sprintf("ReadBytes"),
		}
	}
}

func (t *BufReader) ReadOctetString(length int) []byte {
	if t.err != nil {
		return nil
	}

	tmpBuf := make([]byte, length, length)

	n, err := t.buf.Read(tmpBuf)
	if err != nil {
		t.err = &OpError{
			err: err,
			op:  fmt.Sprintf("ReadOctetString"),
		}
	}

	if n != length {
		t.err = &OpError{
			err: fmt.Errorf("ReadOctetString read %d bytes, not equal %d we expected", n, length),
			op:  fmt.Sprintf("ReadOctetString"),
		}
		return nil
	}

	// 定位字符串后面的0
	i := bytes.IndexByte(tmpBuf, 0)
	if i == -1 {
		return tmpBuf
	} else {
		return tmpBuf[:i]
	}
}

func (t *BufReader) Error() error {
	if t.err != nil {
		return t.err
	}

	return nil
}
