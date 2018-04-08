package app

import (
	"github.com/meixinyun/common/cmd/comm-cmd/app/options"
	"github.com/meixinyun/common/pkg/version/verflag"
	"github.com/meixinyun/common/pkg/server"
	"os"
	"fmt"
	utilflag "github.com/meixinyun/common/util/flag"
	"github.com/golang/glog"
	"github.com/meixinyun/common/pkg/version"
	"github.com/meixinyun/common/pkg/server/mock"
	"github.com/spf13/cobra"
)

func NewCmdServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()
	cmd := &cobra.Command{
		Use:"cmd-server",
		Long:"The comnon cmd server is  a example for cobra command test,just for fun",
		Run:func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()
			utilflag.PrintFlags(cmd.Flags())
			stopCh := server.SetupSignalHandler()
			if err := Run(s, stopCh); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

func Run(runOptions *options.ServerRunOptions, stopCh <-chan struct{}) error {
	glog.Infof("Version: %+v", version.Get())
	server, err := mock.CreateMockServer(runOptions)
	if err != nil {
		return err
	}
	return server.Run(stopCh)
}



