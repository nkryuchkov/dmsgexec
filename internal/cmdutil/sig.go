package cmdutil

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// MakeSignalCtx makes a signal context.
func MakeSignalCtx() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan os.Signal)
	signal.Notify(ch, []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}...)

	go func() {
		select {
		case sig := <-ch:
			log.Printf("Received signal %v: closing...", sig)
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	return ctx, cancel
}
