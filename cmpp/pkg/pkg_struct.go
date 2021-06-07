package pkg

import (
	"errors"
	cmpp "github.com/bigwhite/gocmpp"
	cmpputils "github.com/bigwhite/gocmpp/utils"
	"log"
	"math"
	"mock-cmpp-stress-test/utils/cache"
	"time"
)

const (
	defaultRetries = 3
	defaultTimeout = 5 * time.Second
)

type Cmpp2ConnReqPkg struct {
	SrcAddr   string
	AuthSrc   string
	Version   Version
	Timestamp uint32
	Secret    string
	SeqId     uint32
}

type CmppClientManager struct {
	// setting
	Addr     string        // cmpp server address
	UserName string        // cmpp connect username
	password string        // cmpp connect password
	spId     string        // cmpp submit sp_id
	spCode   string        // cmpp submit sp_code
	retries  uint8         // cmpp connect retry times
	timeout  time.Duration // cmpp connect timeout

	Client    *cmpp.Client // cmpp client
	connected bool
	Cache     *cache.Cache
}

func NewClient(version string) (*cmpp.Client, error) {
	v := GetVersion(version)
	if v == InvalidVersion {
		return nil, errors.New("invalid cmpp version")
	}
	return cmpp.NewClient(v), nil
}

var TpUdhiSeq byte = 0x00

func (cm *CmppClientManager) SplitLongSms(content string) [][]byte {
	smsLength := 140
	smsHeaderLength := 6
	smsBodyLen := smsLength - smsHeaderLength
	contentBytes := []byte(content)
	var chunks [][]byte
	num := 1
	if (len(content)) > 140 {
		num = int(math.Ceil(float64(len(content)) / float64(smsBodyLen)))
	}
	if num == 1 {
		chunks = append(chunks, contentBytes)
		return chunks
	}
	tpUdhiHeader := []byte{0x05, 0x00, 0x03, TpUdhiSeq, byte(num)}
	TpUdhiSeq++

	for i := 0; i < num; i++ {
		chunk := tpUdhiHeader
		chunk = append(chunk, byte(i+1))
		smsBodyLen := smsLength - smsHeaderLength
		offset := i * smsBodyLen
		max := offset + smsBodyLen
		if max > len(content) {
			max = len(content)
		}

		chunk = append(chunk, contentBytes[offset:max]...)
		chunks = append(chunks, chunk)
	}
	return chunks
}
