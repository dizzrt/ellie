package config

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/viper"
)

var _ Config = &stdViperConfig{}

type stdViperConfig struct {
	v *viper.Viper

	envPrefix             string
	isAutomaticEnvEnabled bool
	viperOpts             []viper.Option

	validEnvs        []string
	configFileDir    string
	configFileName   string
	configFileSuffix string
}

func NewStdViperConfig(opts ...Option) *stdViperConfig {
	c := &stdViperConfig{
		viperOpts: make([]viper.Option, 0),

		envPrefix:             "",
		isAutomaticEnvEnabled: true,
		validEnvs:             []string{"local", "dev", "test", "ppe", "prod"},

		configFileDir:    "./configs",
		configFileName:   "config",
		configFileSuffix: ".toml",
	}

	for _, opt := range opts {
		opt(c)
	}

	// create viper instance
	c.v = viper.NewWithOptions(c.viperOpts...)
	return c
}

func (c *stdViperConfig) V() *viper.Viper {
	return c.v
}

func (c *stdViperConfig) Load() error {
	var err error
	var conf string

	if c.isAutomaticEnvEnabled {
		if c.envPrefix != "" {
			c.v.SetEnvPrefix(c.envPrefix)
		}

		c.v.AutomaticEnv()
		fmt.Printf("successfully loaded envs with prefix '%s'\n", c.envPrefix)
	}

	// load default config
	if conf, err = c.GetConfigPath(""); err != nil {
		return err
	}

	c.v.SetConfigFile(conf)
	if err = c.v.ReadInConfig(); err != nil {
		return err
	}
	fmt.Printf("successfully loaded config file '%s'\n", conf)

	// load env config
	env := c.v.GetString("ENV")
	if conf, err = c.GetConfigPath(env); err != nil {
		return err
	}

	if err = c.mergeInConfig(conf, env); err != nil {
		return err
	}

	// load local config
	if conf, err = c.GetConfigPath("local"); err != nil {
		return err
	}

	if err = c.mergeInConfig(conf, "local"); err != nil {
		return err
	}

	return nil
}

func (c *stdViperConfig) Get(key string) Value {
	return &stdValue{val: c.v.Get(key)}
}

func (c *stdViperConfig) IsSet(key string) bool {
	return c.v.IsSet(key)
}

func (c *stdViperConfig) Unmarshal(obj any) error {
	return c.v.Unmarshal(obj)
}

func (c *stdViperConfig) UnmarshalKey(key string, obj any) error {
	return c.v.UnmarshalKey(key, obj)
}

func (c *stdViperConfig) GetConfigPath(env string) (string, error) {
	if env != "" && !slices.Contains(c.validEnvs, env) {
		// TODO std error
		return "", fmt.Errorf("invalid env: %s", env)
	}

	envSeparator := ""
	if env != "" {
		envSeparator = "."
	}

	conf := fmt.Sprintf("%s/%s%s%s%s", c.configFileDir, c.configFileName, envSeparator, env, c.configFileSuffix)
	return conf, nil
}

func (c *stdViperConfig) mergeInConfig(conf string, env string) error {
	if _, err := os.Stat(conf); err != nil {
		if !os.IsNotExist(err) {
			return err
		} else if env != "local" {
			fmt.Printf("config file '%s' is not exist.\n", conf)
		}

		return nil
	}

	c.v.SetConfigFile(conf)
	if err := c.v.MergeInConfig(); err != nil {
		return err
	}

	fmt.Printf("successfully loaded config file '%s'\n", conf)
	return nil
}
