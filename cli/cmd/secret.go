package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/gorilla/securecookie"
	"github.com/spf13/cobra"
)

var size int

// secretCmd generates a secret key using securecookie.GenerateRandomKey
var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Generates a hex secret key",
	Long: `Generates a hex secret key for securecookie.GenerateRandomKey. 
	This can be used with fir.WithSessionSecrets to set the session secret key. The default size is 32.
	Please refer to: https://pkg.go.dev/github.com/gorilla/securecookie#New for more information.`,
	Run: func(cmd *cobra.Command, args []string) {
		key := securecookie.GenerateRandomKey(size)
		hexKey := hex.EncodeToString(key)
		fmt.Println(hexKey)
	},
}

func init() {
	rootCmd.AddCommand(secretCmd)
	secretCmd.Flags().IntVarP(&size, "size", "s", 32, "size of the secret key")
}
