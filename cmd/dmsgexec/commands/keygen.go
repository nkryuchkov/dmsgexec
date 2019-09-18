package commands

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/skycoin/skycoin/src/cipher/bip39"
	"github.com/spf13/cobra"

	"github.com/SkycoinProject/dmsgexec"
	"github.com/SkycoinProject/dmsgexec/internal/cmdutil"
)

func init() {
	rootCmd.AddCommand(keygenCmd)
}

var keysFile = cmdutil.DefaultKeysPath()
var seed string

func init() {
	var err error
	if seed, err = GenerateSeed(); err != nil {
		log.Fatalf("failed to generate seed: %v", err)
	}
	keygenCmd.Flags().StringVar(&keysFile, "keys-file", keysFile, "JSON file to write local keys to")
	keygenCmd.Flags().StringVar(&seed, "seed", seed, "seed to generate keys with (randomly generated if unspecified)")
}

var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generates keys file for dmsgexec",
	Run: func(_ *cobra.Command, _ []string) {
		keys, err := dmsgexec.WriteKeys(keysFile, seed)
		if err != nil {
			log.Fatalf("failed to write keys: %v", err)
		}
		b, _ := json.MarshalIndent(keys, "", "\t") //nolint:errcheck
		fmt.Println(keysFile, ":", string(b))
	},
}

// SeedBitSize represents bit size to use for seed.
const SeedBitSize = 128

func GenerateSeed() (string, error) {
	entropy, err := bip39.NewEntropy(SeedBitSize)
	if err != nil {
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}
	return mnemonic, nil
}
