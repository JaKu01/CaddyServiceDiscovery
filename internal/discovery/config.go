package discovery

import "encoding/json"

type CaddyConfig struct {
	ManualRoutes  []ManualRoute `yaml:"routes"`
	TLSConfig     TLSConfig
	CaddyAdminUrl string
}

type ManualRoute struct {
	Domain   string `yaml:"domain"`
	Upstream string `yaml:"upstreamUrl"`
	TLS      bool   `yaml:"tls"`
}

type TLSConfig struct {
	Manual       bool   `mapstructure:"manual"`
	CertFilePath string `mapstructure:"certFilePath"`
	KeyFilePath  string `mapstructure:"keyFilePath"`
}

func (c CaddyConfig) String() string {
	caddyConfigStr, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return ""
	}
	return string(caddyConfigStr)
}
