package config

import (
	"errors"
	"os"
	"os/user"
	"path"

	"github.com/go-ini/ini"
)

const (
	requiredPermissions = 0600
	defaultHostname     = "manifold.co"
	rcFilename          = ".manifoldrc"
)

// ErrMissingHomeDir represents an error when a home directory could not be found
var ErrMissingHomeDir = errors.New("Could not find Home Directory")

// ErrWrongConfigPermissions represents an error when the config permissions
// are incorrect.
var ErrWrongConfigPermissions = errors.New(
	"~/.manifoldrc must be only readable and writable to the current user (0600)")

// ErrExpectedFile represents an error where the manifoldrc file was a
// directory instead of a file.
var ErrExpectedFile = errors.New("Expected ~/.manifoldrc to be a file; found a directory")

// LoadConfig checks if ~/.manifoldrc exists, if it does, it reads it from
// disk.
//
// If it doesn't an empty config with the default values is returned.
//
// If the file cannot be read, or it has the incorrect permissions an error is
// returned.
func LoadConfig() (*Config, error) {
	rcpath, err := RCPath()
	if err != nil {
		return nil, err
	}

	err = checkPermissions(rcpath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	cfg := &Config{
		Hostname:  defaultHostname,
		AuthToken: "",
	}

	if os.IsNotExist(err) {
		return cfg, nil
	}

	err = ini.MapTo(cfg, rcpath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func checkPermissions(path string) error {
	src, err := os.Stat(path)
	if err != nil {
		return err
	}

	if src.IsDir() {
		return ErrExpectedFile
	}

	if fMode := src.Mode(); fMode.Perm() != requiredPermissions {
		return ErrWrongConfigPermissions
	}

	return nil
}

// RCPath returns the absolute path to the ~/.manifoldrc file
func RCPath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	if u.HomeDir == "" {
		return "", ErrMissingHomeDir
	}

	return path.Join(u.HomeDir, rcFilename), nil
}

// Config represents the configuration which is stored inside a ~/.manifoldrc
// file in ini format.
type Config struct {
	Hostname  string `ini:"hostname"`
	AuthToken string `ini:"auth_token"`
}

// Write writes the contents of the Config struct to ~/.manifoldrc and sets the
// appropriate permissions.
func (c *Config) Write() error {
	rcpath, err := RCPath()
	if err != nil {
		return err
	}

	err = checkPermissions(rcpath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	f, err := os.OpenFile(rcpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, requiredPermissions)
	if err != nil {
		return err
	}

	// Finish writing file before we close, due to race condition inside go-ini
	defer func() {
		f.Sync()
		f.Close()
	}()

	cfg := ini.Empty()
	err = ini.ReflectFrom(cfg, c)
	if err != nil {
		return err
	}

	_, err = cfg.WriteTo(f)
	return err
}
