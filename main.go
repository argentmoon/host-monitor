package main

import (
	"os"
	"os/signal"
	"syscall"

	. "github.com/argentmoon/host-monitor/log"
	"github.com/argentmoon/host-monitor/monitor"
)

func main() {
	GLog.Warn("启动监视器...")
	monitorMgr := monitor.NewMonitorMgr()
	monitorMgr.Load()
	monitorMgr.Start()

	// 阻塞
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
	monitorMgr.CleanAll()
}
