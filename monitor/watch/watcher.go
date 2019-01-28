package watch

import (
	"time"
)

type Watcher interface {
	Name() string
	Host() string
	IsLive() bool
	FreqInSec() time.Duration
}
