package watch

import (
	"time"
)

// Watcher 监测者接口
type Watcher interface {
	Name() string
	Host() string
	IsLive() bool
	WatchType() string
	FreqInSec() time.Duration
}
