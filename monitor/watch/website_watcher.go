package watch

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/argentmoon/host-monitor/log"
)

var httpClient = &http.Client{
	Timeout: time.Second * 20,
	Transport: &http.Transport{
		MaxIdleConns:    100,
		IdleConnTimeout: time.Second * 90,
	},
}

// WebsiteWatcher 网站可用性监测
type WebsiteWatcher struct {
	name       string
	host       string
	httpMethod string // http method: GET, POST ...
	freq       time.Duration
}

// NewWebsiteWatcher 创建新的WebsiteWatcher
func NewWebsiteWatcher(name, host, httpMethod string, freq time.Duration) (websiteWatcher *WebsiteWatcher) {
	log.GLog.Debugf("NewWebsiteWatcher:name:%v, host:%v, httpMethod:%v", name, host, httpMethod)
	return &WebsiteWatcher{
		name:       name,
		host:       host,
		httpMethod: strings.ToUpper(httpMethod),
		freq:       freq,
	}
}

// Name 主机名
func (w *WebsiteWatcher) Name() string {
	return w.name
}

// IsLive 网站是否可用
func (w *WebsiteWatcher) IsLive() (live bool) {
	req, err := http.NewRequest(w.httpMethod, w.host, nil)
	if err != nil {
		log.GLog.Debugf("访问%v，%v错误：%v", w.name, w.host, err)
		return false
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.GLog.Debugf("访问%v，%v错误：%v", w.name, w.host, err)
		return false
	}

	log.GLog.Debugf("访问%v，%v结果：statuscode = %v, status = %v", w.name, w.host, resp.StatusCode, resp.Status)

	return resp.StatusCode == 200 || resp.StatusCode == 403
}

// FreqInSec 监测频率（秒）
func (w *WebsiteWatcher) FreqInSec() time.Duration {
	return w.freq
}

// Host 主机地址
func (w *WebsiteWatcher) Host() string {
	return w.host
}

// WatchType 监测类型
func (w *WebsiteWatcher) WatchType() string {
	return fmt.Sprintf("Http %v监测", w.httpMethod)
}
