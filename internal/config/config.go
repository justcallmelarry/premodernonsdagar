package config

// Config holds application configuration
type Config struct {
	// List of players who opted out from individual statistics
	OptOutList []string
	// Email of the site maintainer
	MaintainerEmail string
	// Server configuration
	Server struct {
		Port string
		Host string
	}
	// Path to log directory
	LogDir string
}

var appConfig Config

// InitConfig initializes the application configuration
func InitConfig() {
	// Set default configuration
	appConfig = Config{
		OptOutList:      []string{"John Doe", "Jane Smith"},
		MaintainerEmail: "test@example.com",
		LogDir:          "logs",
	}

	appConfig.Server.Port = "8080"
	appConfig.Server.Host = "localhost"
}

// GetConfig returns the application configuration
func GetConfig() *Config {
	// If config is not initialized yet, initialize it with defaults
	if appConfig.MaintainerEmail == "" {
		InitConfig()
	}
	return &appConfig
}

// SetOptOutList sets the list of players who opted out
func SetOptOutList(optOutList []string) {
	appConfig.OptOutList = optOutList
}

// AddPlayerToOptOutList adds a player to the opt-out list
func AddPlayerToOptOutList(playerName string) {
	// Check if player is already in the opt-out list
	for _, name := range appConfig.OptOutList {
		if name == playerName {
			return
		}
	}
	appConfig.OptOutList = append(appConfig.OptOutList, playerName)
}

// RemovePlayerFromOptOutList removes a player from the opt-out list
func RemovePlayerFromOptOutList(playerName string) {
	newList := []string{}
	for _, name := range appConfig.OptOutList {
		if name != playerName {
			newList = append(newList, name)
		}
	}
	appConfig.OptOutList = newList
}
