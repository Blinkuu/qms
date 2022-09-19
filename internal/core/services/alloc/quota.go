package alloc

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/Blinkuu/qms/pkg/alloclimit"
)

type quota struct {
	Namespace string                    `yaml:"namespace"`
	Resource  string                    `yaml:"resource"`
	Strategy  alloclimit.StrategyConfig `yaml:"strategy"`
}

type quotaList []quota

func (l *quotaList) String() string {
	bytes, err := yaml.Marshal(*l)
	if err != nil {
		return ""
	}

	return string(bytes)
}

func (l *quotaList) Set(s string) error {
	var result quotaList
	err := yaml.Unmarshal([]byte(s), result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	*l = result

	return nil
}
