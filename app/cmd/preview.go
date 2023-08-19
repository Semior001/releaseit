package cmd

import (
	"context"
	"fmt"
	gengine "github.com/Semior001/releaseit/app/git/engine"
	"github.com/Semior001/releaseit/app/task"
	tengine "github.com/Semior001/releaseit/app/task/engine"
	"os"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/notify"
	"github.com/Semior001/releaseit/app/service/eval"
	"github.com/Semior001/releaseit/app/service/notes"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

// Preview command prints the release notes to stdout.
type Preview struct {
	DataFile     string            `long:"data-file" env:"DATA_FILE" description:"path to the file with release data" required:"true"`
	Extras       map[string]string `long:"extras" env:"EXTRAS" env-delim:"," description:"extra variables to use in the template, will be merged (env primary) with ones in the config file"`
	ConfLocation string            `long:"conf-location" env:"CONF_LOCATION" description:"location to the config file" required:"true"`
}

// Execute prints the release notes to stdout.
func (p Preview) Execute(_ []string) error {
	f, err := os.ReadFile(p.DataFile)
	if err != nil {
		return fmt.Errorf("read data file: %w", err)
	}

	var data struct {
		From         string                     `yaml:"from"`
		To           string                     `yaml:"to"`
		Extras       map[string]string          `yaml:"extras"`
		PullRequests map[string]git.PullRequest `yaml:"pull_requests"`
		Tasks        map[string]task.Ticket     `yaml:"tasks"`
	}

	if err = yaml.Unmarshal(f, &data); err != nil {
		return fmt.Errorf("unmarshal data: %w", err)
	}

	rnbCfg, err := notes.ConfigFromFile(p.ConfLocation)
	if err != nil {
		return fmt.Errorf("read release notes builder config: %w", err)
	}

	evaler := &eval.Evaluator{
		Addon: eval.MultiAddon{
			&eval.Git{Engine: gengine.Unsupported{}},
			&eval.Task{Tracker: &tengine.Tracker{Interface: &tengine.InterfaceMock{
				GetFunc: func(ctx context.Context, id string) (task.Ticket, error) {
					if t, ok := data.Tasks[id]; ok {
						return t, nil
					}
					return task.Ticket{}, fmt.Errorf("task %s not found", id)
				},
				ListFunc: func(ctx context.Context, ids []string) ([]task.Ticket, error) {
					var result []task.Ticket
					for _, id := range ids {
						result = append(result, data.Tasks[id])
					}
					return result, nil
				},
			}}},
		},
	}

	rnb, err := notes.NewBuilder(rnbCfg, evaler, lo.Assign(data.Extras, p.Extras))
	if err != nil {
		return fmt.Errorf("prepare release notes builder: %w", err)
	}

	rn, err := rnb.Build(context.Background(), notes.BuildRequest{
		From:      data.From,
		To:        data.To,
		ClosedPRs: lo.Values(data.PullRequests),
	})
	if err != nil {
		return fmt.Errorf("build release notes: %w", err)
	}

	wr := &notify.WriterNotifier{
		Writer: os.Stdout,
		Name:   "stdout",
	}

	if err = wr.Send(context.Background(), rn); err != nil {
		return fmt.Errorf("print release notes: %w", err)
	}

	return nil
}
