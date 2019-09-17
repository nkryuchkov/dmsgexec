package commands

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/skywire/pkg/util/pathutil"
	"github.com/spf13/cobra"

	"github.com/SkycoinProject/dmsgexec"
	"github.com/SkycoinProject/dmsgexec/internal/cmdutil"
)

var conf = dmsgexec.DefaultServerConfig(cipher.GenerateKeyPair())
var authFile = filepath.Join(pathutil.HomeDir(), ".dmsgssh/whitelist.json")

func init() {
	rootCmd.Flags().Var(&conf.PubKey, "pk", "public key of dmsgexec-server (random if not specified)")
	rootCmd.Flags().Var(&conf.SecKey, "sk", "secret key of dmsgexec-server (random if not specified)")
	rootCmd.Flags().StringVar(&conf.DmsgDisc, "dmsg-disc", conf.DmsgDisc, "address of dmsg discovery to use")
	rootCmd.Flags().Uint16Var(&conf.DmsgPort, "dmsg-port", conf.DmsgPort, "dmsg port to listen on")
	rootCmd.Flags().StringVar(&conf.CLINet, "cli-net", conf.CLINet, "network used for CLI")
	rootCmd.Flags().StringVar(&conf.CLIAddr, "cli-addr", conf.CLIAddr, "address used for CLI")
	rootCmd.Flags().StringVar(&authFile, "auth-file", authFile, "JSON file that contains whitelisted public keys")
}

var rootCmd = &cobra.Command{
	Use: "dmsgexec-server",
	Short: "Server for dmsgexec",
	Run: func(*cobra.Command, []string) {
		ctx, cancel := cmdutil.MakeSignalCtx()
		defer cancel()

		fmt.Println("DMSG ADDRESS:", dmsg.Addr{PK: conf.PubKey, Port: conf.DmsgPort})

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
