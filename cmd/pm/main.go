package main

import (
	"fmt"

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

	rootCmd.Execute()
}
