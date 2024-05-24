package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/bosh-prometheus/cf_exporter/models"
	clients "github.com/cloudfoundry-community/go-cf-clients-helper/v2"
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
	defer httpres.Body.Close()

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

func (s SessionExt) GetTasks() ([]models.Task, error) {
	res := []models.Task{}
	_, _, err := s.V3().MakeListRequest(ccv3.RequestParams{
		RequestName:  "GetTasks",
		Query:        []ccv3.Query{LargeQuery, TaskActiveStates},
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
	defer resp.Body.Close()
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

func (s SessionExt) ListDroplets() ([]models.Droplet, error) {
	client := s.Raw()
	url := fmt.Sprintf("%s/v3/droplets", s.V3().CloudControllerURL)
	var droplets []models.Droplet

	for {
		req, err := client.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code %d on request %s", resp.StatusCode, url)
		}

		var data struct {
			Pagination struct {
				Next struct {
					Href string `json:"href"`
				} `json:"next"`
			} `json:"pagination"`
			Resources []models.Droplet `json:"resources"`
		}
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&data)
		if err != nil {
			return nil, err
		}

		droplets = append(droplets, data.Resources...)

		if data.Pagination.Next.Href == "" {
			break
		}
		nextURL, err := getNextURL(url, data.Pagination.Next.Href)
		if err != nil {
			return nil, err
		}
		url = nextURL
	}

	return droplets, nil
}

func getNextURL(currentURL, nextHref string) (string, error) {
	parsedNext, err := url.Parse(nextHref)
	if err != nil {
		return "", err
	}
	if parsedNext.IsAbs() {
		return parsedNext.String(), nil
	}

	parsedCurrent, err := url.Parse(currentURL)
	if err != nil {
		return "", err
	}

	resolved := parsedCurrent.ResolveReference(parsedNext)
	return resolved.String(), nil
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
