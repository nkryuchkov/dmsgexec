package commands

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/SkycoinProject/dmsgexec"
	"github.com/SkycoinProject/dmsgexec/internal/cmdutil"
)

var keysFile = cmdutil.DefaultKeysPath()
var authFile = cmdutil.DefaultAuthPath()
var conf = dmsgexec.DefaultServerConfig(dmsgexec.Keys{})

func init() {
	rootCmd.Flags().StringVar(&keysFile, "keys-file", keysFile, "JSON file that contains local keys")
	rootCmd.Flags().StringVar(&authFile, "auth-file", authFile, "JSON file that contains whitelisted public keys")

	rootCmd.Flags().StringVar(&conf.DmsgDisc, "dmsg-disc", conf.DmsgDisc, "address of dmsg discovery to use")
	rootCmd.Flags().Uint16Var(&conf.DmsgPort, "dmsg-port", conf.DmsgPort, "dmsg port to listen on")
	rootCmd.Flags().StringVar(&conf.CLINet, "cli-net", conf.CLINet, "network used for CLI")
	rootCmd.Flags().StringVar(&conf.CLIAddr, "cli-addr", conf.CLIAddr, "address used for CLI")

}

var rootCmd = &cobra.Command{
	Use:   "dmsgexec-server",
	Short: "Server for dmsgexec",
	Run: func(*cobra.Command, []string) {
		ctx, cancel := cmdutil.MakeSignalCtx()
		defer cancel()

		keys, err := dmsgexec.ReadKeys(keysFile)
		if err != nil {
			fmt.Println("Run 'dmsgexec keygen' to generate keys file.")
			log.Fatalf("failed to read keys file: %v", err)
		}
		conf.Keys = keys

		whitelist, err := dmsgexec.NewJsonFileWhiteList(authFile)
		if err != nil {
			log.Fatalf("failed to init whitelist: %v", err)
		}

		server := dmsgexec.NewServer(whitelist, conf)
		fmt.Println("Process ended:", server.Serve(ctx))
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
