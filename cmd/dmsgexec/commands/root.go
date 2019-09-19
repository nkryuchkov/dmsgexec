package commands

import (
	"fmt"
	"log"
	"net"

	"github.com/spf13/cobra"

	"github.com/SkycoinProject/dmsgexec"
	"github.com/SkycoinProject/dmsgexec/internal/cmdutil"
)

var cliNet = dmsgexec.DefaultCLINet
var cliAddr = dmsgexec.DefaultCLIAddr
var cmdIn dmsgexec.Command

func init() {
	rootCmd.PersistentFlags().StringVar(&cliNet, "cli-net", cliNet, "network to use for dialing to dmsgexec-server")
	rootCmd.PersistentFlags().StringVar(&cliAddr, "cli-addr", cliAddr, "address to use for dialing to dmsgexec-server")

	cmdIn.DstPort = dmsgexec.DefaultDmsgPort
	rootCmd.Flags().Var(&cmdIn.DstPK, "pk", "remote public key")
	rootCmd.Flags().Uint16Var(&cmdIn.DstPort, "port", cmdIn.DstPort, "remote port")
	rootCmd.Flags().StringVar(&cmdIn.Name, "cmd", cmdIn.Name, "command to execute")
	rootCmd.Flags().StringArrayVar(&cmdIn.Args, "arg", cmdIn.Args, "argument for command")
}

var rootCmd = &cobra.Command{
	Use:   "dmsgexec",
	Short: "Run commands over dmsg",
	Run: func(*cobra.Command, []string) {
		cmdutil.SignalDial(cliNet, cliAddr, func(conn net.Conn) {
			out, err := dmsgexec.Exec(conn, cmdIn)
			if err != nil {
				log.Fatalf("execution failed: %v", err)
			}
			if out[len(out)-1] != '\n' {
				out = append(out, '\n')
			}
			fmt.Print(string(out))
		})
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
