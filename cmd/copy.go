/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"

	registry "github.com/akakream/MultiPlatform2IPFS/internal/registry"
	"github.com/spf13/cobra"
)

var (
	// ErrImageIDRequired is error for when image ID is required
	ErrImageRequired = errors.New("image is required")
	// ErrOnlyOneArgumentRequired is error for when one argument only is required
	ErrOnlyOneArgumentRequired = errors.New("only one argument is required")
)

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "A brief description of your command",
	Long: `copy multi-platform image to IPFS. For example:
MultiPlatform2IPFS copy busybox:latest .`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return ErrImageRequired
		}
		if len(args) != 1 {
			return ErrOnlyOneArgumentRequired
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		registry.CopyImage(args[0])
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// copyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// copyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
