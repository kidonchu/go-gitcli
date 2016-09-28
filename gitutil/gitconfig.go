package gitutil

import (
	"fmt"
	"os"

	git "github.com/libgit2/git2go"
)

var (
	globalConfig *git.Config
	localConfig  *git.Config
)

// initConfig initializes git config object from local .git/config file
// scope can be either "global" or "local"
func initConfig() error {

	if configPath, err := git.ConfigFindGlobal(); err == nil {
		globalConfig, _ = git.OpenOndisk(nil, configPath)
	}

	if dir, err := os.Getwd(); err == nil {
		configPath := fmt.Sprintf("%s/.git/config", dir)
		localConfig, _ = git.OpenOndisk(nil, configPath)
	}

	if !hasConfig() {
		return fmt.Errorf("Could not find any git config file")
	}

	return nil
}

func hasConfig() bool {
	return globalConfig != nil || localConfig != nil
}

// ConfigString finds string value from git config
func ConfigString(name string) (string, error) {

	// Check if config has already been initialized
	if !hasConfig() {
		if err := initConfig(); err != nil {
			return "", err
		}
	}

	if result, err := localConfig.LookupString(name); err == nil {
		return result, nil
	}

	// fall back to global config if not found in local config
	if result, err := globalConfig.LookupString(name); err == nil {
		return result, nil
	}

	return "", fmt.Errorf("No result found in git config files for `%s`", name)
}

// ConfigInt32 finds string value from git config
func ConfigInt32(name string) (int32, error) {

	// Check if config has already been initialized
	if !hasConfig() {
		if err := initConfig(); err != nil {
			return 0, err
		}
	}

	if result, err := localConfig.LookupInt32(name); err == nil {
		return result, nil
	}

	// fall back to global config if not found in local config
	if result, err := globalConfig.LookupInt32(name); err == nil {
		return result, nil
	}

	return 0, fmt.Errorf("No result found in git config files for `%s`", name)
}

// SetConfigString sets string value to git config
func SetConfigString(name, value string) error {

	// Check if config has already been initialized
	if !hasConfig() {
		if err := initConfig(); err != nil {
			return err
		}
	}

	err := localConfig.SetString(name, value)
	if err != nil {
		return fmt.Errorf("Unable to set string config `%s` to `%s`\n%+v", name, value, err)
	}

	return nil
}

// SetConfigInt32 sets int32 value to git config
func SetConfigInt32(name string, value int32) error {

	// Check if config has already been initialized
	if !hasConfig() {
		if err := initConfig(); err != nil {
			return err
		}
	}

	err := localConfig.SetInt32(name, value)
	if err != nil {
		return fmt.Errorf("No result found in gitconfig for `%s`", name)
	}

	return nil
}

// DeleteConfig deletes config
func DeleteConfig(name string) error {

	// Check if config has already been initialized
	if !hasConfig() {
		if err := initConfig(); err != nil {
			return err
		}
	}

	globalConfig.Delete(name)
	localConfig.Delete(name)

	return nil
}
