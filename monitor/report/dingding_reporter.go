package report

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/argentmoon/host-monitor/log"
)

type DingdingReporter struct {
	report_url string
}

func NewDingdingReporter(report_url string) *DingdingReporter {
	GLog.Infof("NewDingdingReporter:%v", report_url)
	return &DingdingReporter{report_url: report_url}
}

func (r *DingdingReporter) Report(msg string) (err error) {
	postMsg := fmt.Sprintf("{ \"msgtype\": \"text\", \"text\": {\"content\": \"%s\"}}", msg)

	_, err = http.Post(r.report_url, "application/json; charset=utf-8", strings.NewReader(postMsg))
	if err != nil {
		GLog.Error(err)
	}
	return
}
