package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/wpajqz/linker"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "version of linker",
	Aliases: []string{"v"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stdout, "Server:")
		fmt.Fprintln(os.Stdout, " Version:             ", linker.VERSION)
		fmt.Fprintln(os.Stdout, " Go version:          ", runtime.Version())
		fmt.Fprintln(os.Stdout, " Minimum Go Version   ", linker.MINIMUM_GO_VERSION)
	},
}
