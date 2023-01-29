package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/notify"
	"github.com/Semior001/releaseit/app/service/notes"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

// Preview command prints the release notes to stdout.
type Preview struct {
	DataFile     string            `long:"data-file" env:"DATA_FILE" description:"path to the file with release data" required:"true"`
	Extras       map[string]string `long:"extras" env:"EXTRAS" env-delim:"," description:"extra variables to use in the template, will be merged (env primary) with ones in the config file"`
	ConfLocation string            `long:"conf_location" env:"CONF_LOCATION" description:"location to the config file" required:"true"`
}

// Execute prints the release notes to stdout.
func (p Preview) Execute(_ []string) error {
	f, err := os.ReadFile(p.DataFile)
	if err != nil {
		return fmt.Errorf("read data file: %w", err)
	}

	var data struct {
		Version      string            `yaml:"version"`
		FromSHA      string            `yaml:"from_sha"`
		ToSHA        string            `yaml:"to_sha"`
		Extras       map[string]string `yaml:"extras"`
		PullRequests []git.PullRequest `yaml:"pull_requests"`
	}

	if err = yaml.Unmarshal(f, &data); err != nil {
		return fmt.Errorf("unmarshal data: %w", err)
	}

	rnbCfg, err := notes.ConfigFromFile(p.ConfLocation)
	if err != nil {
		return fmt.Errorf("read release notes builder config: %w", err)
	}

	rnb, err := notes.NewBuilder(rnbCfg, lo.Assign(data.Extras, p.Extras))
	if err != nil {
		return fmt.Errorf("prepare release notes builder: %w", err)
	}

	rn, err := rnb.Build(notes.BuildRequest{
		FromSHA:   data.FromSHA,
		ToSHA:     data.ToSHA,
		ClosedPRs: data.PullRequests,
	})
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
