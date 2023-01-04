package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/notify"
	"github.com/Semior001/releaseit/app/service/notes"
	"gopkg.in/yaml.v3"
)

// Preview command prints the release notes to stdout.
type Preview struct {
	Version  string `long:"version" env:"VERSION" description:"version to be released" required:"true"`
	DataFile string `long:"data-file" env:"DATA_FILE" description:"path to the file with release data" required:"true"`

	ConfLocation string            `long:"conf_location" env:"CONF_LOCATION" description:"location to the config file" required:"true"`
	Extras       map[string]string `long:"extras" env:"EXTRAS" env-delim:"," description:"extra variables to use in the template"`
}

// Execute prints the release notes to stdout.
func (p Preview) Execute(_ []string) error {
	builder, err := notes.NewBuilder(p.ConfLocation, p.Extras)
	if err != nil {
		return fmt.Errorf("prepare release notes builder: %w", err)
	}

	data, err := os.ReadFile(p.DataFile)
	if err != nil {
		return fmt.Errorf("read data file: %w", err)
	}

	var prs []git.PullRequest
	if err = yaml.Unmarshal(data, &prs); err != nil {
		return fmt.Errorf("unmarshal data: %w", err)
	}

	rn, err := builder.Build(p.Version, prs)
	if err != nil {
		return fmt.Errorf("build release notes: %w", err)
	}

	wr := &notify.WriterNotifier{
		Writer: os.Stdout,
		Name:   "stdout",
	}

	if err = (wr).Send(context.Background(), "", rn); err != nil {
		return fmt.Errorf("print release notes: %w", err)
	}

	return nil
}
