package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/argentmoon/host-monitor/utils"

	. "github.com/argentmoon/host-monitor/log"
	"github.com/argentmoon/host-monitor/monitor"
)

func main() {
	isAdministrator := utils.IsAdministrator()
	if !isAdministrator {
		GLog.Fatal("需要以管理员模式运行！")
	}

	GLog.Info("启动监视器...")
	monitorMgr := monitor.NewMonitorMgr()
	monitorMgr.Load()
	monitorMgr.Start()

	// 阻塞
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
	monitorMgr.CleanAll()
}
