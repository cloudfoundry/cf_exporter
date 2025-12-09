package clients

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	ccWrapper "code.cloudfoundry.org/cli/v8/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/v8/api/router"
	routerWrapper "code.cloudfoundry.org/cli/v8/api/router/wrapper"
	"code.cloudfoundry.org/cli/v8/api/uaa"
	"code.cloudfoundry.org/cli/v8/api/uaa/constant"
	uaaWrapper "code.cloudfoundry.org/cli/v8/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/v8/util/configv3"
)

// Session - wraps the available clients from CF cli
type Session struct {
	clientV3  *ccv3.Client
	clientUAA *uaa.Client
	rawClient *RawClient

	// To call tcp routing with this router
	routerClient *router.Client

	// netClient permit to access to networking policy api
	// Deprecated: cfnetworking-cli-api is deprecated, use Raw() client instead
	netClient *RawClient

	config      Config
	configStore *configv3.Config
}

// NewSession -
func NewSession(c Config) (s *Session, err error) {
	c.Endpoint = strings.TrimSuffix(c.Endpoint, "/")
	if c.User == "" && c.CFClientID == "" {
		return nil, fmt.Errorf("couple of user/password or uaa_client_id/uaa_client_secret must be set")
	}
	if c.User != "" && c.CFClientID == "" {
		c.CFClientID = "cf"
		c.CFClientSecret = ""
	}
	if c.Password == "" && c.CFClientID != "cf" && c.CFClientSecret != "" {
		c.User = ""
	}
	s = &Session{
		config: c,
	}
	config := &configv3.Config{
		ConfigFile: configv3.JSONConfig{
			ConfigVersion:        3,
			Target:               c.Endpoint,
			UAAOAuthClient:       c.CFClientID,
			UAAOAuthClientSecret: c.CFClientSecret,
			SkipSSLValidation:    c.SkipSslValidation,
		},
		ENV: configv3.EnvOverride{
			CFUsername: c.User,
			CFPassword: c.Password,
			BinaryName: c.BinName,
		},
	}
	s.configStore = config
	uaaClientId := c.UaaClientID
	uaaClientSecret := c.UaaClientSecret
	if uaaClientId == "" {
		uaaClientId = c.CFClientID
		uaaClientSecret = c.CFClientSecret
	}
	configUaa := &configv3.Config{
		ConfigFile: configv3.JSONConfig{
			ConfigVersion:        3,
			UAAOAuthClient:       uaaClientId,
			UAAOAuthClientSecret: uaaClientSecret,
			SkipSSLValidation:    c.SkipSslValidation,
		},
	}

	err = s.init(config, configUaa, c)
	if err != nil {
		return nil, fmt.Errorf("error when creating clients: %s", err.Error())
	}
	return s, nil
}

// V3 Give access to api cf v3 (complete and always up to date, thanks to cli v7 team)
func (s *Session) V3() *ccv3.Client {
	return s.clientV3
}

// UAA Give access to api uaa (incomplete)
func (s *Session) UAA() *uaa.Client {
	return s.clientUAA
}

// TCPRouter Give access to TCP Routing api
func (s *Session) TCPRouter() *router.Client {
	return s.routerClient
}

// Networking Give access to networking policy api
// Deprecated: cfnetworking-cli-api is deprecated. Use Raw() client with V3().NetworkPolicyV1() endpoint
// Example: session.Raw().Get(session.V3().NetworkPolicyV1() + "/networking/v1/external/policies")
func (s *Session) Networking() *RawClient {
	return s.netClient
}

// Raw Give an http client which pass authorization header to call api(s) directly
func (s *Session) Raw() *RawClient {
	return s.rawClient
}

// ConfigStore Give config store for client which need access token
func (s *Session) ConfigStore() *configv3.Config {
	return s.configStore
}

func (s *Session) init(config *configv3.Config, configUaa *configv3.Config, configSess Config) error {
	ccWrappersV3 := []ccv3.ConnectionWrapper{}
	authWrapperV3 := ccWrapper.NewUAAAuthentication(nil, config)

	ccWrappersV3 = append(ccWrappersV3, authWrapperV3)
	ccWrappersV3 = append(ccWrappersV3, ccWrapper.NewRetryRequest(config.RequestRetryCount()))
	if s.IsDebugMode() {
		ccWrappersV3 = append(ccWrappersV3, ccWrapper.NewRequestLogger(NewRequestLogger()))
	}

	ccClientV3 := ccv3.NewClient(ccv3.Config{
		AppName:            config.BinaryName(),
		AppVersion:         config.BinaryVersion(),
		JobPollingTimeout:  config.OverallPollingTimeout(),
		JobPollingInterval: config.PollingInterval(),
		Wrappers:           ccWrappersV3,
	})

	ccClientV3.TargetCF(ccv3.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	root, _, err := ccClientV3.GetRoot()
	if err != nil {
		return fmt.Errorf("could not fetch api root information: %s", err)
	}

	// create an uaa client with cf_username/cf_password or client_id/client secret
	// to use it for authenticate requests
	uaaClient := uaa.NewClient(config)
	_, _ = uaaClient.GetAPIVersion()
	uaaAuthWrapper := uaaWrapper.NewUAAAuthentication(nil, configUaa)
	uaaClient.WrapConnection(uaaAuthWrapper)
	uaaClient.WrapConnection(uaaWrapper.NewRetryRequest(config.RequestRetryCount()))
	err = uaaClient.SetupResources(root.UAA(), root.Login())
	if err != nil {
		return fmt.Errorf("error setup resource uaa: %s", err)
	}

	// -------------------------
	// try connecting with pair given on uaa to retrieve access token and refresh token
	var accessToken string
	var refreshToken string
	if config.CFUsername() != "" {
		accessToken, refreshToken, err = uaaClient.Authenticate(map[string]string{
			"username": config.CFUsername(),
			"password": config.CFPassword(),
		}, "", constant.GrantTypePassword)
		config.SetUAAGrantType(string(constant.GrantTypePassword))
	} else if config.UAAOAuthClient() != "cf" {
		accessToken, refreshToken, err = uaaClient.Authenticate(map[string]string{
			"client_id":     config.UAAOAuthClient(),
			"client_secret": config.UAAOAuthClientSecret(),
		}, "", constant.GrantTypeClientCredentials)
		config.SetUAAGrantType(string(constant.GrantTypeClientCredentials))
	}
	if err != nil {
		return fmt.Errorf("error when authenticate on cf: %s", err)
	}
	if accessToken == "" {
		return fmt.Errorf("a pair of username/password or a pair of client_id/client_secret muste be set")
	}

	config.SetAccessToken(fmt.Sprintf("bearer %s", accessToken))
	config.SetRefreshToken(refreshToken)

	// -------------------------
	// assign uaa client to request wrappers
	uaaAuthWrapper.SetClient(uaaClient)
	authWrapperV3.SetClient(uaaClient)
	// -------------------------

	// store client in the sessions
	s.clientV3 = ccClientV3
	// -------------------------

	// -------------------------
	// Create uaa client with given admin client_id only if user give it
	if configUaa.UAAOAuthClient() != "" {
		uaaClientSess := uaa.NewClient(configUaa)

		uaaAuthWrapperSess := uaaWrapper.NewUAAAuthentication(nil, configUaa)
		uaaClientSess.WrapConnection(uaaAuthWrapperSess)
		uaaClientSess.WrapConnection(uaaWrapper.NewRetryRequest(config.RequestRetryCount()))
		err = uaaClientSess.SetupResources(root.UAA(), root.Login())
		if err != nil {
			return fmt.Errorf("error setup resource uaa: %s", err)
		}

		var accessTokenSess string
		var refreshTokenSess string
		if configUaa.UAAOAuthClient() == "cf" {
			accessTokenSess, refreshTokenSess, err = uaaClientSess.Authenticate(map[string]string{
				"username": config.CFUsername(),
				"password": config.CFPassword(),
			}, "", constant.GrantTypePassword)
			configUaa.SetUAAGrantType(string(constant.GrantTypePassword))
		} else {
			accessTokenSess, refreshTokenSess, err = uaaClientSess.Authenticate(map[string]string{
				"client_id":     configUaa.UAAOAuthClient(),
				"client_secret": configUaa.UAAOAuthClientSecret(),
			}, "", constant.GrantTypeClientCredentials)
			configUaa.SetUAAGrantType(string(constant.GrantTypeClientCredentials))
		}

		if err != nil {
			return fmt.Errorf("error when authenticate on uaa: %s", err)
		}
		if accessTokenSess == "" {
			return fmt.Errorf("a pair of pair of uaa_client_id/uaa_client_secret muste be set")
		}
		configUaa.SetAccessToken(fmt.Sprintf("bearer %s", accessTokenSess))
		configUaa.SetRefreshToken(refreshTokenSess)
		s.clientUAA = uaaClientSess
		uaaAuthWrapperSess.SetClient(uaaClientSess)
	}
	// -------------------------

	// -------------------------
	// Create networking policy client using raw client (cfnetworking-cli-api is deprecated)
	// The netClient now points to a configured raw client for network policy API calls
	authWrapperNet := ccWrapper.NewUAAAuthentication(nil, config)
	authWrapperNet.SetClient(uaaClient)
	netWrappers := []ccv3.ConnectionWrapper{
		authWrapperNet,
		NewRetryRequest(config.RequestRetryCount()),
	}
	if s.IsDebugMode() {
		netWrappers = append(netWrappers, ccWrapper.NewRequestLogger(NewRequestLogger()))
	}
	s.netClient = NewRawClient(RawClientConfig{
		ApiEndpoint:       s.clientV3.NetworkPolicyV1(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	}, netWrappers...)
	// -------------------------

	// -------------------------
	// Create raw http client with uaa client authentication to make raw request
	authWrapperRaw := ccWrapper.NewUAAAuthentication(nil, config)
	authWrapperRaw.SetClient(uaaClient)
	rawWrappers := []ccv3.ConnectionWrapper{
		authWrapperRaw,
		NewRetryRequest(config.RequestRetryCount()),
	}
	if s.IsDebugMode() {
		rawWrappers = append(rawWrappers, ccWrapper.NewRequestLogger(NewRequestLogger()))
	}
	s.rawClient = NewRawClient(RawClientConfig{
		ApiEndpoint:       config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	}, rawWrappers...)
	// -------------------------

	// -------------------------
	// Create router client for tcp routing
	routerConfig := router.Config{
		AppName:    config.BinaryName(),
		AppVersion: config.BinaryVersion(),
		ConnectionConfig: router.ConnectionConfig{
			DialTimeout:       config.DialTimeout(),
			SkipSSLValidation: config.SkipSSLValidation(),
		},
		RoutingEndpoint: root.Routing(),
	}

	routerWrappers := []router.ConnectionWrapper{}

	rAuthWrapper := routerWrapper.NewUAAAuthentication(uaaClient, config)
	errorWrapper := routerWrapper.NewErrorWrapper()
	retryWrapper := newRetryRequestRouter(config.RequestRetryCount())

	routerWrappers = append(routerWrappers, rAuthWrapper, retryWrapper, errorWrapper)
	routerConfig.Wrappers = routerWrappers

	s.routerClient = router.NewClient(routerConfig)
	// -------------------------

	return nil
}

func (s *Session) IsDebugMode() bool {
	return s.config.Debug
}
