package main

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/spf13/cobra"
)

func main() {
	log.SetHandler(text.New(os.Stderr))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	addr := "0.0.0.0:" + port
	key := ""
	cert := ""

	rootCmd := &cobra.Command{
		Use:   "operator",
		Short: "Operator facilitates WebRTC signaling",
		Run: func(cmd *cobra.Command, args []string) {

			tlsConfig, err := httpsSettings(key, cert)
			if err != nil {
				println("failed to load tls config:", err.Error())
				os.Exit(1)
			}

			if err := runHTTP(addr, tlsConfig); err != nil {
				log.WithField("bind-addr", addr).
					WithError(err).
					Fatal("failed to run http server")
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&addr, "bind-addr", "", addr, "the address to bind")
	rootCmd.PersistentFlags().StringVarP(&key, "key", "", key, "path to key file")
	rootCmd.PersistentFlags().StringVarP(&cert, "cert", "", cert, "path to cert file")

	err := rootCmd.Execute()
	if err != nil {
		rootCmd.Usage()
	}
}
