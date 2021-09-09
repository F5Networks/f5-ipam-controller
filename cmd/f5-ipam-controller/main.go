package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/F5Networks/f5-ipam-controller/pkg/controller"
	"github.com/F5Networks/f5-ipam-controller/pkg/manager"
	"github.com/F5Networks/f5-ipam-controller/pkg/orchestration"
	log "github.com/F5Networks/f5-ipam-controller/pkg/vlogger"
	clog "github.com/F5Networks/f5-ipam-controller/pkg/vlogger/console"
	flag "github.com/spf13/pflag"
)

const (
	DefaultProvider = manager.F5IPAMProvider
)

var (
	// To be set by build
	version   string
	buildInfo string

	// Flag sets and supported flags
	flags          *flag.FlagSet
	globalFlags    *flag.FlagSet
	basicProvFlags *flag.FlagSet
	ibFlags        *flag.FlagSet

	// Global
	logLevel *string
	orch     *string
	provider *string

	// Default Provider
	iprange *string

	// Infoblox
	ibHost       *string
	ibVersion    *string
	ibPort       *string
	ibUsername   *string
	ibPassword   *string
	ibLabelMap   *string
	printVersion *bool
	ibNetView    *string
	credsDir     *string
)

func init() {
	flags = flag.NewFlagSet("main", flag.ContinueOnError)
	globalFlags = flag.NewFlagSet("Global", flag.ContinueOnError)
	basicProvFlags = flag.NewFlagSet("Default Provider", flag.ContinueOnError)
	ibFlags = flag.NewFlagSet("Infoblox", flag.ContinueOnError)

	//Flag terminal wrapping
	var err error
	var width int
	fd := int(os.Stdout.Fd())
	if terminal.IsTerminal(fd) {
		width, _, err = terminal.GetSize(fd)
		if nil != err {
			width = 0
		}
	}

	// Global flags
	logLevel = globalFlags.String("log-level", "INFO", "Optional, logging level.")
	orch = globalFlags.String("orchestration", "",
		"Required, orchestration that the controller is running in.")
	provider = globalFlags.String("ipam-provider", DefaultProvider,
		"Required, the IPAM system that the controller will interface with.")

	iprange = basicProvFlags.String("ip-range", "",
		"Optional, the Default Provider needs iprange to build pools of IP Addresses")

	printVersion = globalFlags.Bool("version", false,
		"Optional, print version and exit.")

	// Infoblox flags
	ibHost = ibFlags.String("infoblox-grid-host", "",
		"Required for infoblox, the grid manager host IP.")
	ibVersion = ibFlags.String("infoblox-wapi-version", "",
		"Required for infoblox, the Web API version.")
	ibPort = ibFlags.String("infoblox-wapi-port", "443",
		"Optional for infoblox, the Web API port.")
	ibUsername = ibFlags.String("infoblox-username", "",
		"Required for infoblox, the login username.")
	ibPassword = ibFlags.String("infoblox-password", "",
		"Required for infoblox, the login password.")
	ibLabelMap = ibFlags.String("infoblox-labels", "",
		"Required for mapping the infoblox's dnsview and cidr to IPAM labels")
	ibNetView = ibFlags.String("infoblox-netview", "",
		"Required for allocation of IP addresses")
	credsDir = ibFlags.String("credentials-directory", "",
		"Optional, directory that contains the Infoblox username, password and/or wapi-port, grid-host "+
			"files. To be used instead of username, password, and/or wapi-port, grid-host arguments.")

	globalFlags.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "  Global:\n%s\n", globalFlags.FlagUsagesWrapped(width))
	}

	basicProvFlags.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "  Default Provider:\n%s\n", basicProvFlags.FlagUsagesWrapped(width))
	}

	ibFlags.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "  Infoblox Provider:\n%s\n", ibFlags.FlagUsagesWrapped(width))
	}

	flags.AddFlagSet(globalFlags)
	flags.AddFlagSet(basicProvFlags)
	flags.AddFlagSet(ibFlags)

	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s\n", os.Args[0])
		globalFlags.Usage()
		basicProvFlags.Usage()
		ibFlags.Usage()
	}
}

func verifyArgs() error {
	log.RegisterLogger(
		log.LL_MIN_LEVEL, log.LL_MAX_LEVEL, clog.NewConsoleLogger())

	if ll := log.NewLogLevel(*logLevel); nil != ll {
		log.SetLogLevel(*ll)
	} else {
		return fmt.Errorf("Unknown log level requested: %v\n"+
			"    Valid log levels are: DEBUG, INFO, WARNING, ERROR, CRITICAL", logLevel)
	}

	if len(*orch) == 0 {
		return fmt.Errorf("orchestration is required")
	}

	*orch = strings.ToLower(*orch)
	*provider = strings.ToLower(*provider)
	if len(*iprange) == 0 && *provider == DefaultProvider {
		return fmt.Errorf("IP Range not provider for Provider: %v", DefaultProvider)
	}
	*iprange = strings.Trim(*iprange, "\"")
	*iprange = strings.Trim(*iprange, "'")

	if *provider == manager.InfobloxProvider {
		if len(*credsDir) == 0 {
			if len(*ibHost) == 0 || len(*ibVersion) == 0 {
				return fmt.Errorf("missing required Infoblox parameter")
			} else if len(*ibUsername) == 0 || len(*ibPassword) == 0 {
				return fmt.Errorf("missing Infoblox credentials")
			} else if len(*ibLabelMap) == 0 {
				return fmt.Errorf("missing Infoblox Labels")
			} else if len(*ibNetView) == 0 {
				return fmt.Errorf("missing Infoblox NetView")
			}
		}
	}

	return nil
}

func getCredentials() error {
	// Get infoblox credentials
	if len(*credsDir) > 0 {
		var username, password, infobloxURL, port, appendSlash string
		var err error

		if !strings.HasSuffix(*credsDir, "/") {
			appendSlash = "/"
		}
		username = *credsDir + appendSlash + "username"
		password = *credsDir + appendSlash + "password"
		infobloxURL = *credsDir + appendSlash + "grid-host"
		port = *credsDir + appendSlash + "wapi-port"

		setField := func(field *string, filename, fieldType string) error {
			fileBytes, readErr := ioutil.ReadFile(filename)
			if readErr != nil {
				log.Debugf("No %s in credentials directory, falling back to CLI argument", fieldType)
				if len(*field) == 0 {
					return fmt.Errorf("Infoblox %s not specified", fieldType)
				}
			} else {
				*field = strings.TrimSpace(string(fileBytes))
			}
			return nil
		}

		err = setField(ibUsername, username, "username")
		if err != nil {
			return err
		}
		err = setField(ibPassword, password, "password")
		if err != nil {
			return err
		}
		err = setField(ibHost, infobloxURL, "grid-host")
		if err != nil {
			return err
		}
		err = setField(ibPort, port, "wapi-port")
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	err := flags.Parse(os.Args)
	if nil != err {
		os.Exit(1)
	}

	if *printVersion {
		fmt.Printf("Version: %s\nBuild: %s\n", version, buildInfo)
		os.Exit(0)
	}
	err = verifyArgs()
	if nil != err {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		flags.Usage()
		os.Exit(1)
	}
	err = getCredentials()
	if nil != err {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		flags.Usage()
		os.Exit(1)
	}
	log.Infof("[INIT] Starting: F5 IPAM Controller - Version: %s, BuildInfo: %s", version, buildInfo)

	orcr := orchestration.NewOrchestrator()
	if orcr == nil {
		log.Error("Unable to create IPAM Client")
		os.Exit(1)
	}
	mgrParams := manager.Params{
		Provider: *provider,
	}
	switch *provider {
	case manager.F5IPAMProvider:
		mgrParams.IPAMManagerParams = manager.IPAMManagerParams{Range: *iprange}
	case manager.InfobloxProvider:
		mgrParams.InfobloxParams = manager.InfobloxParams{
			Host:       *ibHost,
			Version:    *ibVersion,
			Port:       *ibPort,
			Username:   *ibUsername,
			Password:   *ibPassword,
			IbLabelMap: *ibLabelMap,
			NetView:    *ibNetView,
		}
	}
	mgr, err := manager.NewManager(mgrParams)
	if err != nil {
		log.Errorf("Unable to initialize manager: %v", err)
		os.Exit(1)
	}
	stopCh := make(chan struct{})
	ctlr := controller.NewController(
		controller.Spec{
			Orchestrator: orcr,
			Manager:      mgr,
			StopCh:       stopCh,
		},
	)
	ctlr.Start()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals

	ctlr.Stop()
	log.Infof("Exiting - signal %v\n", sig)
	close(stopCh)
}
