package commands

import (
	"fmt"
	"log"
	"net"

	"github.com/spf13/cobra"

	"github.com/SkycoinProject/dmsgexec"
	"github.com/SkycoinProject/dmsgexec/internal/cmdutil"
)

var cmdIn dmsgexec.Command
var cliNet = dmsgexec.DefaultCLINet
var cliAddr = dmsgexec.DefaultCLIAddr

func init() {
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
		ctx, cancel := cmdutil.MakeSignalCtx()
		defer cancel()

		conn, err := net.Dial(cliNet, cliAddr)
		if err != nil {
			log.Fatalf("failed to dial to dmsgexec-server: %v", err)
		}

		go func() {
			<-ctx.Done()
			_ = conn.Close() //nolint:errcheck
		}()

		out, err := dmsgexec.Exec(conn, cmdIn)
		if err != nil {
			log.Fatalf("execution failed: %v", err)
		}
		if out[len(out)-1] != '\n' {
			out = append(out, '\n')
		}
		fmt.Print(out)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
