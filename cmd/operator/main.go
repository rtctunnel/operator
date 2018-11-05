package main

import (
	"net"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "operator",
	Short: "Operator facilitates WebRTC signaling",
}

func main() {
	log.SetHandler(text.New(os.Stderr))

	addr := "localhost:8000"
	rootCmd.PersistentFlags().StringVarP(&addr, "bind-addr", "", addr, "the address to bind")

	li, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithField("bind-addr", addr).
			WithError(err).
			Fatal("failed to start listener")
	}

	err = runHTTP(li)
	if err != nil {
		log.WithField("bind-addr", addr).
			WithError(err).
			Fatal("failed to run http server")
	}
}
