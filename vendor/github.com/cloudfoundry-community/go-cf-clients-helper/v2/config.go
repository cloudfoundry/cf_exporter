package clients

// config -
type Config struct {
	// Cloud foundry api endoint
	Endpoint string
	// cloud foundry user to call api (CFClientID can be set instead)
	User string
	// cloud foundry password to call api (CFClientID can be set instead)
	Password string
	// cloud foundry client id to call api or to identify user (User can be set in addition)
	// If empty and user set this value will be set as `cf`
	CFClientID string
	// cloud foundry client secret to call api or to identify user (user can be set in addition)
	CFClientSecret string
	// uaa client id to call uaa api
	UaaClientID string
	// uaa client secret to call uaa api
	UaaClientSecret string
	// Skip ssl verification
	SkipSslValidation bool
	// Show debug trace
	Debug bool
	// Binary name that will be used as user-agent
	BinName string
}
