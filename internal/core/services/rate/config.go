package rate

type Config struct {
	Backend string  `mapstructure:"backend"`
	Quotas  []quota `mapstructure:"quotas"`
}
