package pkg

import (
	"fmt"
	cmpp "github.com/bigwhite/gocmpp"
	"math"
	"mock-cmpp-stress-test/utils/cache"
	"strconv"
	"time"
)

const (
	defaultRetries = 3
	defaultTimeout = 5 * time.Second
)

// cmpp client
type CmppClientManager struct {
	// setting
	Addr     string        // cmpp server address
	Version  cmpp.Type     // cmpp version
	UserName string        // cmpp connect username
	Password string        // cmpp connect password
	SpId     string        // cmpp submit sp_id
	SpCode   string        // cmpp submit sp_code
	Retries  uint          // cmpp connect retry times
	Timeout  time.Duration // cmpp connect timeout

	Client    *cmpp.Client // cmpp client
	Connected bool
	Cache     *cache.Cache
}

// cmpp server
type CmppServerManager struct {
	// setting
	Addr          string    // cmpp client address
	Version       cmpp.Type // cmpp version
	heartbeat     time.Duration
	maxNoRespPkgs int32

	ConnMap map[string]*Conn
	UserMap map[string]*Conn
}

type Conn struct {
	UserName string // cmpp connect auth username
	password string // cmpp connect auth password
	spId     string // cmpp submit sp_id
	spCode   string // cmpp submit sp_code
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

func GetMsgId(spId string, seqId uint32) (uint64, error) {
	now := time.Now()
	month, _ := strconv.ParseInt(fmt.Sprintf("%d", now.Month()), 10, 32)
	day := now.Day()
	hour := now.Hour()
	min := now.Minute()
	sec := now.Second()
	spIdInt, _ := strconv.ParseInt(spId, 10, 32)
	binaryStr := fmt.Sprintf("%04b%05b%05b%06b%06b%022b%016b", month, day, hour, min, sec, spIdInt, seqId)
	msgId, err := strconv.ParseUint(binaryStr, 2, 64)
	if err != nil {
		return 0, err
	}
	return msgId, nil
}
