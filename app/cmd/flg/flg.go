package flg

import (
	"fmt"
	"regexp"

	"github.com/Semior001/releaseit/app/service"
)

// ServiceGroup defines parameters to initialize application service.
type ServiceGroup struct {
	SquashCommitRx string      `long:"squash-commit-rx" env:"SQUASH_COMMIT_RX" description:"regexp to match squash commits" default:"^squash:(.?)+$"`
	Engine         EngineGroup `group:"engine" namespace:"engine" env-namespace:"ENGINE"`
	Notify         NotifyGroup `group:"notify" namespace:"notify" env-namespace:"NOTIFY"`
}

// Build creates a new service.Service instance.
func (s ServiceGroup) Build() (*service.Service, error) {
	eng, err := s.Engine.Build()
	if err != nil {
		return nil, fmt.Errorf("prepare engine: %w", err)
	}

	notif, err := s.Notify.Build()
	if err != nil {
		return nil, fmt.Errorf("prepare notifier: %w", err)
	}

	rnb, err := s.Notify.ReleaseNotesBuilder()
	if err != nil {
		return nil, fmt.Errorf("prepare release notes builder: %w", err)
	}

	rx, err := regexp.Compile(s.SquashCommitRx)
	if err != nil {
		return nil, fmt.Errorf("compile squash commit regexp: %w", err)
	}

	return &service.Service{
		Engine:                eng,
		ReleaseNotesBuilder:   rnb,
		Notifier:              notif,
		SquashCommitMessageRx: rx,
	}, nil
}
