// Package main is an entrypoint for application
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/Semior001/releaseit/app/cmd"
	"github.com/hashicorp/logutils"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	Changelog cmd.Changelog `command:"changelog"    description:"build release notes for a pair of commits"`
	Preview   cmd.Preview   `command:"preview"      description:"preview release notes with data read from file"`
	Debug     bool          `long:"dbg" env:"DEBUG" description:"turn on debug mode"`
}

var version = "unknown"

func getVersion() string {
	v, ok := debug.ReadBuildInfo()
	if !ok || v.Main.Version == "(devel)" {
		return version
	}
	return v.Main.Version
}

func main() {
	_, _ = fmt.Fprintf(os.Stderr, "releaseit, version: %s\n", getVersion())

	p := flags.NewParser(&opts, flags.Default)
	p.CommandHandler = func(cmd flags.Commander, args []string) error {
		setupLog(opts.Debug)

		err := cmd.Execute(args)
		if err != nil {
			log.Printf("[ERROR] failed to run command: %+v", err)
		}

		return err
	}

	if _, err := p.Parse(); err != nil {
		var flagsErr *flags.Error

		if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}

		os.Exit(1)
	}
}

func setupLog(dbg bool) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: "INFO",
		Writer:   os.Stderr,
	}

	logFlags := log.Ldate | log.Ltime

	if dbg {
		logFlags = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
		filter.MinLevel = "DEBUG"
	}

	log.SetFlags(logFlags)
	log.SetOutput(filter)
}
