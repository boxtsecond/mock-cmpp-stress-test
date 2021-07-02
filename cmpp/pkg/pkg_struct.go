package pkg

import (
	"context"
	cmpp "github.com/bigwhite/gocmpp"
	"sync"
	"time"
)

const (
	defaultRetries = 3
	defaultTimeout = 5 * time.Second
)

// cmpp client
type CmppClientManager struct {
	// setting
	Addr     string    // cmpp test address
	Version  cmpp.Type // cmpp version
	UserName string    // cmpp connect username
	Password string    // cmpp connect password
	SpId     string    // cmpp submit sp_id
	SpCode   string    // cmpp submit sp_code
	//Retries            uint          // cmpp connect retry times
	Timeout            time.Duration // cmpp connect timeout
	ActiveTestInterval time.Duration // cmpp connect timeout

	Connected    bool
	ConnErrCount uint
	Ctx          context.Context
	cancel       context.CancelFunc

	Client          *cmpp.Client // cmpp client
	Cmpp2SubmitChan chan *cmpp.Cmpp2SubmitReqPkt
	Cmpp3SubmitChan chan *cmpp.Cmpp3SubmitReqPkt
}

// cmpp test
type CmppServerManager struct {
	// setting
	Addr          string    // cmpp client address
	Version       cmpp.Type // cmpp version
	heartbeat     time.Duration
	maxNoRespPkgs int32
	ConnMap       *sync.Map //map[string]*cmpp.Conn // 连接池
	UserMap       *sync.Map //[string]*Conn // 用户map

	SubmitSeqId <-chan uint16
	SubmitDone  chan<- struct{}
}

type Conn struct {
	UserName string // cmpp connect auth username
	password string // cmpp connect auth password
	spId     string // cmpp submit sp_id
	spCode   string // cmpp submit sp_code
}
