# zaphandler

zaphandler will help to create [slog handler](https://pkg.go.dev/golang.org/x/exp/slog#Handler) using [zap logger](https://pkg.go.dev/go.uber.org/zap) [![Go Version](https://img.shields.io/github/go-mod/go-version/chanchal1987/zaphandler.svg)](https://github.com/chanchal1987/zaphandler) [![Go Reference](https://pkg.go.dev/badge/go.mrchanchal.com/zaphandler.svg)](https://pkg.go.dev/go.mrchanchal.com/zaphandler) [![License](https://badgen.net/github/license/chanchal1987/zaphandler)](https://github.com/chanchal1987/zaphandler/blob/main/LICENSE) [![Report Card](https://goreportcard.com/badge/go.mrchanchal.com/zaphandler)](https://goreportcard.com/report/go.mrchanchal.com/zaphandler)

## Example:

    package main

    import (
        "go.mrchanchal.com/zaphandler"
        "go.uber.org/zap"
        "golang.org/x/exp/slog"
    )

    func main() {
        zapL, _ := zap.NewDevelopment()
        defer zapL.Sync()

        logger := slog.New(zaphandler.New(zapL))

        logger.Info("sample log message", "field1", "value1", "field2", 33)
    }

[Go Playground](https://go.dev/play/p/J7anFkS4KKO)