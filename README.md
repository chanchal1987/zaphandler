# zaphandler

zaphandler will help to create [slog handler](https://pkg.go.dev/log/slog#Handler) using [zap logger](https://pkg.go.dev/go.uber.org/zap) [![Go Reference](https://pkg.go.dev/badge/go.mrchanchal.com/zaphandler.svg)](https://pkg.go.dev/go.mrchanchal.com/zaphandler) [![Report Card](https://goreportcard.com/badge/go.mrchanchal.com/zaphandler)](https://goreportcard.com/report/go.mrchanchal.com/zaphandler)

## Example:

package main

    import (
        "log/slog"

        "go.mrchanchal.com/zaphandler"
        "go.uber.org/zap"
    )

    func main() {
        zapL, _ := zap.NewDevelopment()
        defer zapL.Sync()

        logger := slog.New(zaphandler.New(zapL))

        logger.Info("sample log message", "field1", "value1", "field2", 33)
    }

[Go Playground](https://go.dev/play/p/R57ypgVRB8c?v=gotip)