package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "odyc-cli",
	Short: "CLI tool with handy commands for Odyc.js developers",
	Long:  `Generate code from sprite, or do similar actions making your life with Odyc.js easier.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Logf(3, "Welcome to Odyc.js CLI!")
		log.Info("Add --help to learn how to use this command")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
