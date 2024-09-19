package main

import (
	"github.com/OpenTestSolar/testtool-golang-ginkgo/cmd/build"
	"github.com/OpenTestSolar/testtool-golang-ginkgo/cmd/discover"
	"github.com/OpenTestSolar/testtool-golang-ginkgo/cmd/execute"

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
