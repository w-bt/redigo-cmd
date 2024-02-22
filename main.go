package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"crm-redigo-cmd/pkg/logger"
)

var (
	version = "0.0.1"
	commit  = "n/a"
)

func main() {
	registerStackDumpReceiver()
	cli := newCLI()
	cli.Version = fmt.Sprintf("%s (Commit: %s)", version, commit)
	cli.Execute()
}

func registerStackDumpReceiver() {
	sigChan := make(chan os.Signal)
	go func() {
		stacktrace := make([]byte, 32768)
		for range sigChan {
			length := runtime.Stack(stacktrace, true)
			logger.Error("Stack Trace Dump")
			logger.Errorf(string(stacktrace[:length]))
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
}
