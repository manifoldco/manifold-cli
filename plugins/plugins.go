package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
)

const (
	directory              = ".manifold"
	permissions            = 0700
	pluginPrefix           = "manifold-cli-"
	pluginConfigFilename   = ".config.json"
	pluginConfigPermission = 0644
	slash                  = "/"
)

// ErrFailedToRead is returned when you the plugins directory cannot be read
var ErrFailedToRead = errors.New("Failed to read plugins directory")

// ErrMissingHomeDir represents an error when a home directory could not be found
var ErrMissingHomeDir = errors.New("Could not find Home Directory")

// ErrBadPluginRepository represents an error when the plugin url is invalidu
var ErrBadPluginRepository = errors.New("Invalid repository URL")

// Path returns the plugin directory's path
func Path() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	if u.HomeDir == "" {
		return "", ErrMissingHomeDir
	}

	return path.Join(u.HomeDir, directory), nil
}

// Executable returns the executable path
func Executable(cmd string) (string, error) {
	pluginsDir, err := Path()
	if err != nil {
		return "", err
	}

	// TODO: we need to infer the architecture being user
	// and call the proper binary in the plugin repository
	return path.Join(pluginsDir, cmd, "bin", strings.Replace(cmd, pluginPrefix, "", 1)), nil
}

// List returns a list of available plugin commands
func List() ([]string, error) {
	var plugs []string

	// Construct the plugins directory path
	pluginsDir, err := Path()
	if err != nil {
		return plugs, err
	}

	// Create the directory if it does not exist
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		_ = os.Mkdir(pluginsDir, permissions)
		return plugs, nil
	}

	// Read the contents of the directory
	files, err := ioutil.ReadDir(pluginsDir)
	if err != nil {
		return plugs, ErrFailedToRead
	}
	for _, f := range files {
		plugs = append(plugs, f.Name())
	}

	return plugs, nil
}

// NormalizeURL returns a normalized git url
func NormalizeURL(pluginURL string) string {
	u, err := url.Parse(pluginURL)
	if err != nil {
		return ""
	}
	if strings.HasSuffix(u.Path, slash) {
		u.Path = u.Path[:len(u.Path)-1]
	}
	pluginURL = fmt.Sprintf("git@%s%s", u.Host, u.Path)
	if strings.Count(pluginURL, slash) == 2 {
		pluginURL = strings.Replace(pluginURL, slash, ":", 1)
	}
	return pluginURL
}

// DeriveName returns the name of the plugin based on git repo url
func DeriveName(pluginURL string) string {
	repoURL := NormalizeURL(pluginURL)
	repoURL = repoURL[strings.Index(repoURL, "/")+1:]
	return strings.Replace(repoURL, ".git", "", 1)
}

// Help executes the plugin's help command
func Help(cmd string) error {
	// List installed plugins
	plugs, err := List()
	if err != nil {
		return err
	}

	for _, p := range plugs {
		if cmd == p {
			// Identify the executable path for the plugin
			binPath, err := Executable(cmd)
			if err != nil {
				return err
			}

			// Construct execution of the plugin binary
			proc := exec.Command(binPath, "--help")
			proc.Stdout = os.Stdout
			proc.Stderr = os.Stderr

			// Execute
			return proc.Run()
		}
	}

	// Plugin not found, no error
	return errors.New("Plugin not found")
}

// Run executes the requested plugin command and redirects stdout
func Run(cmd string) (bool, error) {
	// List installed plugins
	plugs, err := List()
	if err != nil {
		return false, err
	}

	for _, p := range plugs {
		if cmd == strings.Replace(p, pluginPrefix, "", 1) {
			// Identify the executable path for the plugin
			binPath, err := Executable(p)
			if err != nil {
				return false, err
			}

			// Construct execution of the plugin binary
			var pluginArgs []string
			pluginArgs = append(pluginArgs, os.Args[2:]...)
			proc := exec.Command(binPath, pluginArgs...)
			proc.Stdout = os.Stdout
			proc.Stderr = os.Stderr

			// Execute
			err = proc.Run()
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}

	// Plugin not found, no error
	return false, nil
}

// configPath returns the plugin config file path
func configPath(plugin string) (string, error) {
	pluginsDir, err := Path()
	if err != nil {
		return "", err
	}

	return path.Join(pluginsDir, pluginPrefix+plugin, pluginConfigFilename), nil
}

// Config returns the key value configuration found in the plugin directory
func Config(plugin string) (map[string]string, string, error) {
	// TODO: make this map[string]interface to allow for subkeys
	conf := make(map[string]string)

	confPath, err := configPath(plugin)
	if err != nil {
		return conf, "", err
	}

	// Create the .config.json if it doesn't exist
	_, err = os.Stat(confPath)
	if os.IsNotExist(err) {
		_, err := os.Create(confPath)
		return conf, confPath, err
	}

	// Read the config file contents
	contents, err := ioutil.ReadFile(confPath)
	if err != nil || len(contents) < 1 {
		return conf, confPath, err
	}

	err = json.Unmarshal(contents, &conf)
	return conf, confPath, err
}

// SetConfig sets a key on the config map
func SetConfig(plugin, key, value string) (map[string]string, error) {
	conf, confPath, err := Config(plugin)
	if err != nil {
		return conf, err
	}

	conf[key] = value
	confJSON, _ := json.Marshal(conf)
	file, err := os.OpenFile(confPath, os.O_RDWR, pluginConfigPermission)
	defer file.Close()
	if err != nil {
		return conf, err
	}
	_, err = file.Write(confJSON)
	if err != nil {
		return conf, err
	}

	return conf, nil
}
