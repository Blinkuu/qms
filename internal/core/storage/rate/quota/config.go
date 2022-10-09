package quota

type Config struct {
	Algorithm      string `yaml:"algorithm"`
	Unit           string `yaml:"unit"`
	RequestPerUnit int64  `yaml:"requests_per_unit"`
}
