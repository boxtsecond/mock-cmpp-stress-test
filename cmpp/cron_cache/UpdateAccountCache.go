package cron_cache

import (
	"mock-cmpp-stress-test/config"
	"sync"
)

// 定时从配置中读取账号信息，更新server端的账号缓存
var cmppAccountCacheObj *CmppAccountCache

type CmppAccountCache struct {
	lock       sync.RWMutex
	accountMap map[string]*config.CmppServerAuth
}

func init() {
	cmppAccountCacheObj = &CmppAccountCache{
		accountMap: make(map[string]*config.CmppServerAuth),
	}
}

// 获取全部cmpp账户
func GetAllCmppAccount() map[string]*config.CmppServerAuth {
	cmppAccountCacheObj.lock.RLock()
	defer cmppAccountCacheObj.lock.RUnlock()
	return cmppAccountCacheObj.accountMap
}

// 获取指定账户信息
func GetAccountInfo(username string) *config.CmppServerAuth {
	key := username
	cmppAccountCacheObj.lock.RLock()
	defer cmppAccountCacheObj.lock.RUnlock()
	if v, ok := cmppAccountCacheObj.accountMap[key]; ok {
		return v
	}
	return nil
}

func UpdateAccountCache() {
	accountMap := make(map[string]*config.CmppServerAuth)
	cfg := config.ConfigObj.ServerConfig
	for _, auth := range *cfg.Auths {
		accountMap[auth.UserName] = &config.CmppServerAuth{
			UserName: auth.UserName,
			Password: auth.Password,
			SpId:     auth.SpId,
			SpCode:   auth.SpCode,
		}
	}
	cmppAccountCacheObj.accountMap = accountMap
}
