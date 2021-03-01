package app

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (

	// Version is version number which automatically set on build. `git describe --tags`
	// set these value as go build -ldflags option
	Version string
	// Revision is git commit hash which automatically set `git rev-parse --short HEAD` on build.
	Revision string
	// GoVersion stores the runtime version of Go
	GoVersion = runtime.Version()
	// Compiler versions
	Compiler = runtime.Compiler
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "secretary-lab",
	Short: "secretary for lab activities",
	Long:  "secretary for lab activities",
	Version: fmt.Sprintf("reposiTree Version: %s (Revision: %s / GoVersion: %s / Compiler: %s)\n",
		Version, Revision, GoVersion, Compiler),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
