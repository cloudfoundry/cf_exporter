package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	clients "github.com/cloudfoundry-community/go-cf-clients-helper/v2"
	"github.com/cloudfoundry/cf_exporter/v2/models"
	log "github.com/sirupsen/logrus"
)

type SessionExt struct {
	clients.Session
}

func NewSessionExt(config *CFConfig) (*SessionExt, error) {
	conf := clients.Config{
		Endpoint:          config.URL,
		SkipSslValidation: config.SkipSSLValidation,
		CFClientID:        config.ClientID,
		CFClientSecret:    config.ClientSecret,
		User:              config.Username,
		Password:          config.Password,
		BinName:           "cf_exporter",
	}
	session, err := clients.NewSession(conf)
	if err != nil {
		log.Errorf("unable to create cf client: %s", err)
		return nil, err
	}

	return &SessionExt{
		Session: *session,
	}, nil
}

func (s SessionExt) GetInfo() (models.Info, error) {
	responseBody := models.Info{}
	res, httpres, err := s.V3().MakeRequestSendReceiveRaw(
		"GET",
		fmt.Sprintf("%s/v3/info", s.V3().CloudControllerURL),
		http.Header{},
		nil,
	)
	if err != nil {
		return responseBody, err
	}
	defer func() {
		if err := httpres.Body.Close(); err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}()
	if httpres.StatusCode != http.StatusOK {
		return responseBody, fmt.Errorf("http error")
	}

	if err = json.Unmarshal(res, &responseBody); err != nil {
		return responseBody, err
	}
	return responseBody, nil
}

func (s SessionExt) GetApplications() ([]models.Application, error) {
	res := []models.Application{}
	_, _, err := s.V3().MakeListRequest(ccv3.RequestParams{
		RequestName:  "GetApplications",
		Query:        []ccv3.Query{LargeQuery},
		ResponseBody: models.Application{},
		AppendToList: func(item interface{}) error {
			res = append(res, item.(models.Application))
			return nil
		},
	})
	return res, err
}

func TaskStatesQuery(states []string) ccv3.Query {
	normalized := normalizeTaskStates(states)
	return ccv3.Query{
		Key:    ccv3.StatesFilter,
		Values: normalized,
	}
}

func normalizeTaskStates(states []string) []string {
	if len(states) == 0 {
		return append([]string{}, DefaultTaskStates...)
	}
	normalized := make([]string, 0, len(states))
	for _, state := range states {
		trimmed := strings.TrimSpace(state)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, strings.ToUpper(trimmed))
	}
	if len(normalized) == 0 {
		return append([]string{}, DefaultTaskStates...)
	}
	return normalized
}

func (s SessionExt) GetTasks(states []string) ([]models.Task, error) {
	res := []models.Task{}
	_, _, err := s.V3().MakeListRequest(ccv3.RequestParams{
		RequestName:  "GetTasks",
		Query:        []ccv3.Query{LargeQuery, TaskStatesQuery(states)},
		ResponseBody: models.Task{},
		AppendToList: func(item interface{}) error {
			res = append(res, item.(models.Task))
			return nil
		},
	})
	return res, err
}

func (s SessionExt) GetOrganizationQuotas() ([]models.Quota, error) {
	res := []models.Quota{}
	_, _, err := s.V3().MakeListRequest(ccv3.RequestParams{
		RequestName:  "GetOrganizationQuotas",
		Query:        []ccv3.Query{LargeQuery},
		ResponseBody: models.Quota{},
		AppendToList: func(item interface{}) error {
			res = append(res, item.(models.Quota))
			return nil
		},
	})
	return res, err
}

func (s SessionExt) GetSpaceQuotas() ([]models.Quota, error) {
	res := []models.Quota{}
	_, _, err := s.V3().MakeListRequest(ccv3.RequestParams{
		RequestName:  "GetSpaceQuotas",
		Query:        []ccv3.Query{LargeQuery},
		ResponseBody: models.Quota{},
		AppendToList: func(item interface{}) error {
			res = append(res, item.(models.Quota))
			return nil
		},
	})
	return res, err
}

func (s SessionExt) GetEvents(query ...ccv3.Query) ([]models.Event, error) {
	res := []models.Event{}
	_, _, err := s.V3().MakeListRequest(ccv3.RequestParams{
		RequestName:  "GetEvents",
		Query:        query,
		ResponseBody: models.Event{},
		AppendToList: func(item interface{}) error {
			res = append(res, item.(models.Event))
			return nil
		},
	})
	return res, err
}

func (s SessionExt) GetSpaceSummary(guid string) (*models.SpaceSummary, error) {
	client := s.Raw()
	url := fmt.Sprintf("/v2/spaces/%s/summary", guid)
	req, err := client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d on request %s", resp.StatusCode, url)
	}
	res := models.SpaceSummary{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
