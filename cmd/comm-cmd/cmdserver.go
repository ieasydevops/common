package main

import (
	"github.com/meixinyun/common/cmd/comm-cmd/app"
	"github.com/spf13/pflag"
	"fmt"
	"os"
	utilflag "github.com/meixinyun/common/util/flag"
	goflag "flag"
)

func main() {
	command := app.NewCmdServerCommand()
	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}




