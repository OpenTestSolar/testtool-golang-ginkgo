package main

import (
	"ginkgo/cmd/build"
	"ginkgo/cmd/discover"
	"ginkgo/cmd/execute"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := cobra.Command{
		Use: "solar-ginkgo",
	}
	rootCmd.AddCommand(discover.NewCmdDiscover())
	rootCmd.AddCommand(execute.NewCmdExecute())
	rootCmd.AddCommand(build.NewCmdBuild())
	_ = rootCmd.Execute()
}
