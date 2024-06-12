package build

import (
	ginkgoBuilder "ginkgo/pkg/builder"

	"github.com/spf13/cobra"
)

type BuildOptions struct {
	projPath string
}

// NewBuildOptions new build options with default value
func NewBuildOptions() *BuildOptions {
	return &BuildOptions{}
}

// NewCmdBuild create a build command
func NewCmdBuild() *cobra.Command {
	o := NewBuildOptions()
	cmd := cobra.Command{
		Use:   "build",
		Short: "Build testcase",
		Run: func(cmd *cobra.Command, args []string) {
			o.RunBuild(cmd, args)
		},
	}
	cmd.Flags().StringVarP(&o.projPath, "root", "r", "", "Project root path")
	cmd.MarkFlagRequired("root")
	return &cmd
}

func (o *BuildOptions) RunBuild(cmd *cobra.Command, args []string) error {
	err := ginkgoBuilder.Build(o.projPath)
	if err != nil {
		return err
	}
	return nil
}
