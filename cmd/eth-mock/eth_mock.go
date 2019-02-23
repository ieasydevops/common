package eth_mock

import (
	"github.com/meixinyun/common/pkg/util/flags"
)

const (
	clientIdentifier = "eth_mock" // Client identifier to advertise over the network
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	// The app that holds all commands and flags.
	app = flags.NewApp(gitCommit, "the go-ethereum command line interface")
)
