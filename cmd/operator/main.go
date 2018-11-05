package main

import (
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

	err := runHTTP(addr)
	if err != nil {
		log.WithError(err).Fatal("failed to run http server")
	}
}
