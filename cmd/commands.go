/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	registry "github.com/akakream/MultiPlatform2IPFS/internal/registry"
	"github.com/akakream/MultiPlatform2IPFS/server"
	"github.com/akakream/MultiPlatform2IPFS/utils"
)

var (
	// ErrImageIDRequired is error for when image ID is required
	ErrImageRequired = errors.New("image is required")
	// ErrOnlyOneArgumentRequired is error for when one argument only is required
	ErrOnlyOneArgumentRequired = errors.New("only one argument is required")
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start server",
	Long:  `Start server`,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := godotenv.Load(); err != nil {
			panic(err)
		}
		baseURL, err := utils.GetEnv("BASE_URL", "localhost:3000")
		if err != nil {
			panic(err)
		}

		s := server.NewServer(baseURL)
		s.Start()
	},
}

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
		imageNameTag := strings.Split(args[0], ":")
		registry.CopyImage(context.TODO(), imageNameTag[0], imageNameTag[1])
	},
}

func init() {
	serverCmd.PersistentFlags().StringP("port", "p", "3002", "give the port where the server runs")
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(copyCmd)
}
