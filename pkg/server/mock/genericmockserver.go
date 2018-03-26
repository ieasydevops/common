package mock

import (
	"github.com/meixinyun/common/util/wait"
	"github.com/meixinyun/common/cmd/comm-cmd/app/options"
	"log"
)

type GenericMockServer struct {
	options *options.ServerRunOptions
}

func (s *GenericMockServer) Run(stopCh <-chan struct{}) error {
	wait.Until(func() {
		log.Print("Hello , I'm Mock Server online")
	}, s.options.AliveTime, stopCh)
	return nil
}

func CreateMockServer(runOptions *options.ServerRunOptions) (*GenericMockServer, error) {
	server := &GenericMockServer{options:runOptions}
	return server, nil
}