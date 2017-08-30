package plugins

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
	"strings"

	"github.com/manifoldco/manifold-cli/config"
)

const (
	directory    = ".manifold"
	permissions  = 0700
	pluginPrefix = "manifold-cli-"
	slash        = "/"
	goarch       = runtime.GOARCH
	goos         = runtime.GOOS
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
	return path.Join(pluginsDir, cmd, "build", goos+"_"+goarch, "bin/", Shortname(cmd)), nil
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

// Shortname returns the shortened name of the plugin
func Shortname(name string) string {
	return strings.Replace(name, pluginPrefix, "", 1)
}

// NormalizeURL returns a normalized git url
func NormalizeURL(pluginURL string) string {
	u, err := url.Parse(pluginURL)
	if err != nil {
		return pluginURL
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

// Config returns the key value configuration found in the plugin directory
func Config(pluginName string, conf interface{}) error {
	manifoldYaml, err := config.LoadYaml(true)
	if err != nil {
		return err
	}
	return manifoldYaml.GetPlugin(pluginName, conf)
}
