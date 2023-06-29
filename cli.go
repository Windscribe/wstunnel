package main

import (
	"github.com/spf13/cobra"
	"os"
	"wstunnel/proxy"
)

var listenAddress string
var remoteAddress string
var tunnelType int
var mtu int
var logFilePath string
var dev = false

var rootCmd = &cobra.Command{
	Use:   "root",
	Short: "Starts local proxy and connects to server.",
	Long:  "Starts local proxy and sets up connection to the server. At minimum it requires remote server address and log file path.",
	Run: func(cmd *cobra.Command, args []string) {
		proxy.Initialise(dev, logFilePath)
		started := proxy.StartProxy(listenAddress, remoteAddress, tunnelType, mtu)
		if started == false {
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&listenAddress, "listenAddress", "l", ":65479", "Local port for proxy > :65479")
	rootCmd.PersistentFlags().StringVarP(&remoteAddress, "remoteAddress", "r", "", "Wstunnel > wss://$ip:$port/tcp/127.0.0.1/$WS_TUNNEL_PORT  Stunnel > https://$ip:$port")
	_ = rootCmd.MarkPersistentFlagRequired("remoteAddress")
	rootCmd.PersistentFlags().IntVarP(&tunnelType, "tunnelType", "t", 1, "WStunnel > 1 , Stunnel > 2")
	rootCmd.PersistentFlags().IntVarP(&mtu, "mtu", "m", 1500, "1500")
	rootCmd.PersistentFlags().StringVarP(&logFilePath, "logFilePath", "f", "", "Path to log file > file.log")
	_ = rootCmd.MarkPersistentFlagRequired("logFilePath")
	rootCmd.PersistentFlags().BoolVarP(&dev, "dev", "d", false, "Turns on verbose logging.")
}

func main() {
	_, err := rootCmd.ExecuteC()
	if err != nil {
		return
	}
}
