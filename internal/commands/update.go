package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Update() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update packages",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Update packages")
		},
	}
}
