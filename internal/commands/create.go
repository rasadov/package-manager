package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Create() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create packages",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Create packages")
		},
	}
}
