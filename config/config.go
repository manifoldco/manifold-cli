package config

import (
	"errors"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/go-ini/ini"
	"github.com/stripe/stripe-go"
	"gopkg.in/yaml.v2"
)

// Version represents the version of the cli. This variable is updated at build
// time.
var Version = "dev"

// StripePublishableKey facilitates secure transmission of payment values
var StripePublishableKey = "pk_live_A6qSWh1v4SrNnrWSftgDcKFQ"

func init() {
	stripe.LogLevel = 0
	stripe.Key = StripePublishableKey
}

const (
	requiredPermissions = 0600
	defaultHostname     = "manifold.co"
	defaultScheme       = "https"
	defaultAnalytics    = true
	rcFilename          = ".manifoldrc"
	rootPathLength      = 1
	// YamlFilename is where dirprefs are stored
	YamlFilename = ".manifold.yml"
)

// ErrMissingHomeDir represents an error when a home directory could not be found
var ErrMissingHomeDir = errors.New("Could not find Home Directory")

// ErrPluginNotFound is returned if a particular plugin doesn't exist in the yaml when requested
var ErrPluginNotFound = errors.New("Plugin not found")

// ErrWrongConfigPermissions represents an error when the config permissions
// are incorrect.
var ErrWrongConfigPermissions = errors.New(
	"~/.manifoldrc must be only readable and writable to the current user (0600)")

// ErrExpectedFile represents an error where the manifoldrc file was a
// directory instead of a file.
var ErrExpectedFile = errors.New("Expected ~/.manifoldrc to be a file; found a directory")

// Load checks if ~/.manifoldrc exists, if it does, it reads it from disk.
//
// If it doesn't an empty config with the default values is returned.
//
// If the file cannot be read, or it has the incorrect permissions an error is
// returned.
func Load() (*Config, error) {
	rcpath, err := RCPath()
	if err != nil {
		return nil, err
	}

	err = checkPermissions(rcpath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	cfg := &Config{
		Hostname:        defaultHostname,
		AuthToken:       "",
		TransportScheme: defaultScheme,
		Analytics:       defaultAnalytics,
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
	Hostname        string `ini:"hostname"`
	AuthToken       string `ini:"auth_token"`
	TransportScheme string `ini:"scheme"`
	Analytics       bool   `ini:"analytics"`
	Team            string `ini:"team"`
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

// ManifoldYaml represents the standard project config object
type ManifoldYaml struct {
	App     string                 `yaml:"app,omitempty" flag:"app,omitempty"`
	Team    string                 `yaml:"team,omitempty" flag:"team,omitempty"`
	Plugins map[string]interface{} `yaml:"plugins,omitempty"`
	Path    string                 `yaml:"-" json:"-"`
}

// GetPlugin retrieves plugins config for the given plugin name
func (m ManifoldYaml) GetPlugin(name string, conf interface{}) error {
	if m.Plugins == nil {
		return ErrPluginNotFound
	}
	if _, ok := m.Plugins[name]; ok {
		// TODO: Can this just be reflected into the interface?
		str, err := yaml.Marshal(m.Plugins[name])
		if err != nil {
			return errors.New("Invalid configuration")
		}
		err = yaml.Unmarshal(str, conf)
		if err != nil {
			return errors.New("Failed to read configuration")
		}
		return nil
	}
	return ErrPluginNotFound
}

// SavePlugin writes the ManifoldYaml values for a specific plugin name
func (m *ManifoldYaml) SavePlugin(name string, conf interface{}) error {
	if m.Plugins == nil {
		m.Plugins = make(map[string]interface{})
	}
	m.Plugins[name] = conf
	return m.Save()
}

// Save writes the ManifoldYaml values to the file in the struct's Path
// field
func (m *ManifoldYaml) Save() error {
	yml, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(m.Path, yml, requiredPermissions)
	if err != nil {
		return err
	}
	return nil
}

// Remove removes the backing file for this ManifoldYaml
func (m *ManifoldYaml) Remove() error {
	return os.Remove(m.Path)
}

// LoadYaml loads ManifoldYaml. It starts in the current working directory,
// looking for a readable '.manifold.yml' file, and walks up the directory
// hierarchy until it finds one, or reaches the root of the fs.
//
// It returns an empty ManifoldYaml is no '.manifold.yml' files are found.
// It returns an error if a malformed file is found, or if any errors occur
// during file system access.
func LoadYaml(recurse bool) (*ManifoldYaml, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	prefs := &ManifoldYaml{}

	var f *os.File
	for {
		f, err = os.Open(filepath.Join(path, YamlFilename))
		if err != nil {
			if isSystemRoot(path) || !recurse {
				return prefs, nil
			}

			path = filepath.Dir(path)
			continue
		}

		break
	}

	defer f.Close()
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(contents, prefs)
	if err != nil {
		return nil, err
	}

	prefs.Path = f.Name()
	return prefs, nil
}

// isSystemRoot validates if the given path is the root of the system for the
// OS the application is running on.
func isSystemRoot(path string) bool {
	if len(path) != rootPathLength {
		return false
	}

	return os.PathSeparator == path[rootPathLength-1]
}
