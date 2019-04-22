package watch

import (
	"net"
	"time"

	"github.com/argentmoon/host-monitor/log"
)

// PortWatcher 端口可用性监测
type PortWatcher struct {
	name string
	host string
	freq time.Duration
}

// NewPortWatcher 创建新的NewPortWatcher
// host 要求是 host : port 模式
func NewPortWatcher(name, host string, freq time.Duration) (w *PortWatcher) {
	log.GLog.Debugf("New PortWatcher:name:%v, host:%v", name, host)
	return &PortWatcher{
		name: name,
		host: host,
		freq: freq,
	}
}

// Name 主机名
func (w *PortWatcher) Name() string {
	return w.name
}

// IsLive 端口是否可用
func (w *PortWatcher) IsLive() (live bool) {
	d := net.Dialer{Timeout: time.Second * 4}
	conn, err := d.Dial("tcp", w.host) //查看是否连接成功
	if err != nil {
		log.GLog.Debugf("访问%v，%v错误：%v", w.name, w.host, err)
		return false
	}

	log.GLog.Debugf("访问%v，%v成功", w.name, w.host)

	conn.Close()
	return true
}

// FreqInSec 监测频率（秒）
func (w *PortWatcher) FreqInSec() time.Duration {
	return w.freq
}

// Host 主机地址
func (w *PortWatcher) Host() string {
	return w.host
}

// WatchType 监测类型
func (w *PortWatcher) WatchType() string {
	return "端口监测"
}
