package pkg

import (
	"go.uber.org/zap"
	"mock-cmpp-stress-test/utils/cache"
	"mock-cmpp-stress-test/utils/log"
	"time"
)

func (cm *CmppClientManager) Init(version, addr, username, password, spId, spCode string, retryTimes uint8, timeout time.Duration) error {
	var err error
	cm.Client, err = NewClient(version)
	if err != nil {
		return err
	}

	cm.Addr = addr
	cm.UserName = username
	cm.password = password
	cm.retries = retryTimes
	cm.timeout = timeout
	cm.spId = spId
	cm.spCode = spCode
	cm.Cache = (&cache.Cache{}).New(1e4)
	cm.Cache.StartRetry()

	if cm.retries == 0 {
		cm.retries = defaultRetries
	}

	if cm.timeout > defaultTimeout {
		cm.timeout = defaultTimeout
	}

	return nil
}

func (cm *CmppClientManager) Connect() error {
	err := cm.Client.Connect(cm.Addr, cm.UserName, cm.password, cm.timeout)
	if err != nil {
		if cm.retries <= 0 {
			log.Logger.Error("[CmppClient][Connect] Error: ", zap.Error(err))
			return err
		}
		cm.retries -= 1
		return cm.Connect()
	}
	cm.connected = true
	log.Logger.Info("[CmppClient][Connect] Success.", zap.String("Addr", cm.Addr), zap.String("UserName", cm.UserName), zap.String("Password", cm.password))
	return nil
}

func (cm *CmppClientManager) Disconnect() {
	cm.Client.Disconnect()
	cm.connected = false
	log.Logger.Info("[CmppClient][Disconnect] Success.", zap.String("Addr", cm.Addr), zap.String("UserName", cm.UserName), zap.String("Password", cm.password))
}
