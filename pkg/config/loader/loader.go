package loader

import (
	"errors"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

const applicationEnv = "APP_PROFILE"

// LoadSettings represents how to load config to config struct.
type LoadSettings struct {
	ConfigFile *ConfigFileSettings
	EnvPrefix  string
	// Flag to load environment variables (load only if file contains same var as env)
	LoadSysEnv bool
}

type ConfigFileSettings struct {
	ConfigDir      string // path to cgoonfig directory
	FileNamePrefix string // config file prefix
	ConfigType     string // config file type (see viper.SupportedExts)
	// SetEnvPrefix defines a prefix that ENVIRONMENT variables will use.
	// See viper.SetEnvPrefix
}

// Load config from files and system env to app config struct.
// Use 'APP_PROFILE' sys env key to denote app profile config.
// Env variables override file vars (can't set without in file vars defining)
func Load(configStruct interface{}, loadSettings *LoadSettings) error {
	v := viper.New()
	if loadSettings == nil {
		return errors.New("loadSettings is nil")
	}
	if loadSettings.ConfigFile == nil {
		return errors.New("config file settings is nil")
	}
	v.SetConfigType(loadSettings.ConfigFile.ConfigType)
	v.AddConfigPath(getConfigDir(loadSettings))
	if err := mergeCommonConfig(v, loadSettings.ConfigFile); err != nil {
		return err
	}
	profileName := getProfileName(loadSettings, v)
	if err := mergeAppProfileConfig(v, profileName, loadSettings.ConfigFile); err != nil {
		return err
	}
	if loadSettings.LoadSysEnv {
		mergeSystemEnvConfig(v, loadSettings.EnvPrefix)
	}
	if err := v.Unmarshal(configStruct); err != nil {
		log.Printf("unable to decode into struct, %v", err)
		return err
	}
	return nil
}

// merge common(base) config file to config struct
func mergeCommonConfig(v *viper.Viper, fileSettings *ConfigFileSettings) error {
	const baseConfigName = "base"
	v.SetConfigName(resolveFileName(fileSettings.FileNamePrefix, baseConfigName))
	if err := v.MergeInConfig(); err != nil {
		log.Printf("%v", err)
		return err
	}
	log.Printf("loaded %s config (%s)", baseConfigName, v.ConfigFileUsed())
	return nil
}

// merge app profile config file to config struct
func mergeAppProfileConfig(v *viper.Viper, profile string, fileSettings *ConfigFileSettings) error {
	if profile != "" {
		v.SetConfigName(resolveFileName(fileSettings.FileNamePrefix, profile))
		if err := v.MergeInConfig(); err != nil {
			if err, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Printf("config for %s profile not found, %v", profile, err)
			} else {
				log.Printf("%+v", err)
			}
			return err
		}
		log.Printf("loaded %s profile config (%s)", profile, v.ConfigFileUsed())
	}
	return nil
}

// merge system environment variables to config struct
func mergeSystemEnvConfig(v *viper.Viper, envPrefix string) {
	v.SetEnvPrefix(envPrefix)
	v.AllowEmptyEnv(false)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}

func getConfigDir(loadSettings *LoadSettings) string {
	configDir := loadSettings.ConfigFile.ConfigDir
	if configDir == "" {
		configDir = "."
	}
	return configDir
}

func getProfileName(loadSettings *LoadSettings, v *viper.Viper) string {
	var envKey = applicationEnv
	if loadSettings.EnvPrefix != "" {
		envKey = strings.ToUpper(loadSettings.EnvPrefix) + "_" + applicationEnv
	}
	profile := os.Getenv(envKey)
	if profile == "" {
		profile = v.GetString("app.profile")
	}
	return profile
}

func resolveFileName(filePrefix, name string) string {
	if filePrefix == "" {
		return name
	}
	return filePrefix + "_" + name
}
