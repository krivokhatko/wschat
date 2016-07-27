package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"

	"github.com/krivokhatko/wschat/pkg/logger"
	"github.com/krivokhatko/wschat/pkg/server"
)

const (
	cmdUsage string = "Usage: %s [OPTIONS]"

	defaultConfigFile string = "/etc/wschat/wschat.json"
	defaultLogPath    string = ""
	defaultLogLevel   string = "warning"
)

var (
	version      string
	buildDate    string
	flagConfig   string
	flagHelp     bool
	flagLogPath  string
	flagLogLevel string
	flagVersion  bool
	logLevel     int
	err          error
)

func PrintUsage(output io.Writer, usage string) {
	fmt.Fprintf(output, usage, path.Base(os.Args[0]))
	fmt.Fprint(output, "\n\nOptions:\n")

	flag.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(output, "   -%s  %s\n", f.Name, f.Usage)
	})

	os.Exit(2)
}

func PrintVersion(version, buildDate string) {
	fmt.Printf("%s version %s, built on %s\nGo version: %s (%s)\n",
		path.Base(os.Args[0]),
		version,
		buildDate,
		runtime.Version(),
		runtime.Compiler,
	)
}

func init() {
	flag.StringVar(&flagConfig, "c", defaultConfigFile, "configuration file path")
	flag.BoolVar(&flagHelp, "h", false, "display this help and exit")
	flag.StringVar(&flagLogPath, "l", defaultLogPath, "log file path")
	flag.StringVar(&flagLogLevel, "L", defaultLogLevel, "logging level (error, warning, notice, info, debug)")
	flag.BoolVar(&flagVersion, "V", false, "display software version and exit")
	flag.Usage = func() { PrintUsage(os.Stderr, cmdUsage) }
	flag.Parse()

	if flagHelp {
		PrintUsage(os.Stdout, cmdUsage)
	} else if flagVersion {
		PrintVersion(version, buildDate)
		os.Exit(0)
	} else if flagConfig == "" {
		fmt.Fprintf(os.Stderr, "Error: configuration file path is mandatory\n")
		PrintUsage(os.Stderr, cmdUsage)
	}

	if logLevel, err = logger.GetLevelByName(flagLogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid log level `%s'\n", flagLogLevel)
		os.Exit(1)
	}
}

func main() {
	instance := server.NewServer(flagConfig, flagLogPath, logLevel)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				instance.Stop()
				break
			}
		}
	}()

	if err := instance.Run(); err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
