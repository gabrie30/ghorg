package configs

import (
	"log"
	"os"
)

// GhorgIgnoreLocation returns the path of users ghorgignore
func GhorgIgnoreLocation() string {
	return GhorgDir() + "/ghorgignore"
}

// ghorgIgnoreDetected identify if a ghorgignore file exists in users .config/ghorg directory
func ghorgIgnoreDetected() bool {
	_, err := os.Stat(GhorgIgnoreLocation())
	return !os.IsNotExist(err)
}

// GhorgDir returns the ghorg directory path
func GhorgDir() string {
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		return os.Getenv("XDG_CONFIG_HOME") + "/ghorg"
	}

	return HomeDir() + "/.config/ghorg"
}

// HomeDir finds the users home directory
func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error trying to find users home directory")
	}

	return home
}
