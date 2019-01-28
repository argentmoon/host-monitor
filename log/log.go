package log

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var GLog = logrus.New()

func init() {
	f, err := exec.LookPath(os.Args[0])
	if err != nil {
		return
	}

	var path string
	path, err = filepath.Abs(f)
	if err != nil {
		return
	}

	path = filepath.Dir(path)

	var logFilename string = filepath.Join(path, "log.log")
	file, err := os.OpenFile(logFilename, os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("open file error=%s\r\n", err.Error())
		os.Exit(-1)
	}

	//not close log file here, otherwise later gLogger is not usable
	//only close log file after whole go file done
	writers := []io.Writer{
		file,
		os.Stdout,
	}

	fileAndStdoutWriter := io.MultiWriter(writers...)

	GLog.Out = fileAndStdoutWriter

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	GLog.SetReportCaller(true)
	GLog.SetLevel(logrus.WarnLevel)
}
