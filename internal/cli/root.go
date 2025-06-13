package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rh-pos",
	Short: "RH-POS is a Point of Sale application",
	Long: `RH-POS is a modern Point of Sale application built with Go.
It provides features for managing products, transactions, and users.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.rh-pos.yaml)")
}
