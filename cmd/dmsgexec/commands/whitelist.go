package commands

import (
	"fmt"
	"log"
	"math/big"
	"net"
	"sort"

	"github.com/skycoin/dmsg/cipher"
	"github.com/spf13/cobra"

	"github.com/SkycoinProject/dmsgexec"
	"github.com/SkycoinProject/dmsgexec/internal/cmdutil"
)

func init() {
	rootCmd.AddCommand(
		whitelistCmd,
		whitelistAddCmd,
		whitelistRemoveCmd)
}

var whitelistCmd = &cobra.Command{
	Use:   "whitelist",
	Short: "lists all whitelisted public keys",
	Run: func(_ *cobra.Command, _ []string) {
		cmdutil.SignalDial(cliNet, cliAddr, func(conn net.Conn) {
			pks, err := dmsgexec.ViewWhitelist(conn)
			if err != nil {
				log.Fatalf("failed to obtain whitelist: %v", err)
			}
			sort.Slice(pks, func(i, j int) bool {
				var a, b big.Int
				a.SetBytes(pks[i][:])
				b.SetBytes(pks[j][:])
				return a.Cmp(&b) >= 0
			})
			for _, pk := range pks {
				fmt.Println(pk)
			}
		})
	},
}

var pk cipher.PubKey

func init() {
	whitelistAddCmd.Flags().Var(&pk, "pk", "public key of remote")
}

var whitelistAddCmd = &cobra.Command{
	Use:   "whitelist-add",
	Short: "adds a public key to whitelist",
	Run: func(_ *cobra.Command, _ []string) {
		cmdutil.SignalDial(cliNet, cliAddr, func(conn net.Conn) {
			if pk.Null() {
				log.Fatal("cannot add a null public key to the whitelist")
			}
			if err := dmsgexec.WhitelistAdd(conn, pk); err != nil {
				log.Fatalf("failed to add public key '%s' to the whitelist: %v", pk, err)
			}
			fmt.Println("OK")
		})
	},
}

func init() {
	whitelistRemoveCmd.Flags().Var(&pk, "pk", "public key of remote")
}

var whitelistRemoveCmd = &cobra.Command{
	Use:   "whitelist-remove",
	Short: "removes a public key from the whitelist",
	Run: func(_ *cobra.Command, _ []string) {
		cmdutil.SignalDial(cliNet, cliAddr, func(conn net.Conn) {
			if pk.Null() {
				log.Fatal("cannot remove a null public key from the whitelist")
			}
			if err := dmsgexec.WhitelistRemove(conn, pk); err != nil {
				log.Fatalf("failed to remove public key '%s' from the whitelist: %v", pk, err)
			}
			fmt.Println("OK")
		})
	},
}
