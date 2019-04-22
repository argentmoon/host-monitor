package watch

import (
	"time"

	"github.com/argentmoon/host-monitor/log"
	"github.com/paulstuart/ping"
)

// PingWatcher 端口可用性监测
type PingWatcher struct {
	name string
	host string
	freq time.Duration
}

// NewPingWatcher 创建新的NewPingWatcher
func NewPingWatcher(name, host string, freq time.Duration) (w *PingWatcher) {
	log.GLog.Debugf("New PingWatcher:name:%v, host:%v", name, host)
	return &PingWatcher{
		name: name,
		host: host,
		freq: freq,
	}
}

// Name 主机名
func (w *PingWatcher) Name() string {
	return w.name
}

// IsLive host ping是否可用
func (w *PingWatcher) IsLive() (live bool) {
	for i := 0; i < 4; i++ {
		// 有一个ping成功，就认为成功
		if ping.Ping(w.host, 3) {
			log.GLog.Debugf("ping访问%v，%v成功", w.name, w.host)
			return true
		}
	}

	log.GLog.Debugf("ping访问%v，%v失败", w.name, w.host)
	return false
}

// FreqInSec 监测频率（秒）
func (w *PingWatcher) FreqInSec() time.Duration {
	return w.freq
}

// Host 主机地址
func (w *PingWatcher) Host() string {
	return w.host
}

// WatchType 监测类型
func (w *PingWatcher) WatchType() string {
	return "ping监测"
}
