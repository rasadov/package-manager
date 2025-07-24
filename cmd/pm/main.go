package main

import (
	"fmt"

	"github.com/rasadov/package-manager/internal/commands"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "pm",
		Short: "Package Manager",
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Version 1.0.0")
		},
	})

	rootCmd.AddCommand(commands.Create())
	rootCmd.AddCommand(commands.Update())

	rootCmd.Execute()
}
