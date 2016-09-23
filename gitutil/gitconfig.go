package gitutil

import (
	"fmt"

	git "github.com/libgit2/git2go"
)

var config *git.Config

// initConfig initializes git config object from global .gitconfig file
func initConfig() error {
	configPath, err := git.ConfigFindGlobal()
	if err != nil {
		return fmt.Errorf("Global .gitconfig could not be found\n%+v\n", err)
	}

	config, err = git.OpenOndisk(nil, configPath)
	if err != nil {
		return fmt.Errorf("Unable to open `%s`\n%+v\n", configPath, err)
	}

	return nil
}

// ConfigString finds string value from git config
func ConfigString(name string) (string, error) {

	// Check if config has already been initialized
	if config == nil {
		if err := initConfig(); err != nil {
			return "", err
		}
	}

	result, err := config.LookupString(name)
	if err != nil {
		return "", fmt.Errorf("No result found in gitconfig for `%s`", name)
	}

	return result, nil
}

// ConfigInt32 finds string value from git config
func ConfigInt32(name string) (int32, error) {

	// Check if config has already been initialized
	if config == nil {
		if err := initConfig(); err != nil {
			return 0, err
		}
	}

	result, err := config.LookupInt32(name)
	if err != nil {
		return 0, fmt.Errorf("No result found in gitconfig for `%s`", name)
	}

	return result, nil
}

// SetConfigString sets string value to git config
func SetConfigString(name, value string) error {

	// Check if config has already been initialized
	if config == nil {
		if err := initConfig(); err != nil {
			return err
		}
	}

	err := config.SetString(name, value)
	if err != nil {
		return fmt.Errorf("Unable to set string config `%s` to `%s`\n%+v", name, value, err)
	}

	return nil
}

// SetConfigInt32 sets int32 value to git config
func SetConfigInt32(name string, value int32) error {

	// Check if config has already been initialized
	if config == nil {
		if err := initConfig(); err != nil {
			return err
		}
	}

	err := config.SetInt32(name, value)
	if err != nil {
		return fmt.Errorf("No result found in gitconfig for `%s`", name)
	}

	return nil
}

// DeleteConfig deletes config
func DeleteConfig(name string) error {

	// Check if config has already been initialized
	if config == nil {
		if err := initConfig(); err != nil {
			return err
		}
	}

	err := config.Delete(name)
	if err != nil {
		return fmt.Errorf("Unable to delete `%s`in gitconfig", name)
	}

	return nil
}
