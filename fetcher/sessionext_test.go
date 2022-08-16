package fetcher

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"
	"net/http"

	"github.com/bosh-prometheus/cf_exporter/models"
)

const (
	fakeToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NjAyMjQzOTksImV4cCI6MTY5MTc2MDM5OSwiYXVkIjoid3d3LmV4YW1wbGUuY29tIiwic3ViIjoianJvY2tldEBleGFtcGxlLmNvbSIsIkdpdmVuTmFtZSI6IkpvaG5ueSIsIlN1cm5hbWUiOiJSb2NrZXQiLCJFbWFpbCI6Impyb2NrZXRAZXhhbXBsZS5jb20iLCJSb2xlIjpbIk1hbmFnZXIiLCJQcm9qZWN0IEFkbWluaXN0cmF0b3IiXX0.Qh0M7WikXNEH1aHnsp5fGVrV2JKZpFV6OxDlQKX68Jk"
)

func serializeList[T any](obj ...T) string {
	content := serialize(obj)
	response := &ccv3.PaginatedResources{
		ResourcesBytes: []byte(content),
	}
	return serialize(response)
}

func serialize[T any](obj T) string {
	content, err := json.Marshal(obj)
	Ω(err).ShouldNot(HaveOccurred())
	return string(content)
}

var _ = Describe("Extensions", func() {

	var (
		err    error
		server *ghttp.Server
		target *SessionExt
		config *CFConfig
	)

	BeforeEach(func() {
		tokenResponse := fmt.Sprintf(`{"access_token": "%s", "refresh_token": "value"}`, fakeToken)
		server = ghttp.NewServer()

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/"),
				ghttp.RespondWith(http.StatusOK, serialize(ccv3.Info{
					Links: ccv3.InfoLinks{
						Login: resources.APILink{HREF: server.URL()},
						UAA:   resources.APILink{HREF: server.URL()},
					},
				})),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/oauth/token"),
				ghttp.RespondWith(http.StatusOK, tokenResponse),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/oauth/token"),
				ghttp.RespondWith(http.StatusOK, tokenResponse),
			),
		)

		config = &CFConfig{
			URL:          server.URL(),
			ClientID:     "fake",
			ClientSecret: "fake",
		}
		target, err = NewSessionExt(config)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(target).ShouldNot(BeNil())
	})

	AfterEach(func() {
		server.Close()
	})

	Context("fetching applications", func() {
		It("no error occurs", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v3/apps", "per_page=5000"),
					ghttp.RespondWith(http.StatusOK, serializeList(models.Application{
						GUID:  "app1-guid",
						Name:  "app1",
						State: constant.ApplicationStarted,
						Relationships: resources.Relationships{
							constant.RelationshipTypeSpace: resources.Relationship{GUID: "app1-space-guid"},
						},
					})),
				),
			)
			app, err := target.GetApplications()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(app).Should(HaveLen(1))
			Ω(app[0].GUID).Should(Equal("app1-guid"))
			Ω(app[0].Name).Should(Equal("app1"))
		})
	})


	Context("fetching org quotas", func() {
		It("no error occurs", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v3/organization_quotas", "per_page=5000"),
					ghttp.RespondWith(http.StatusOK, serializeList(
						models.Quota{
							GUID: "guid1",
							Name: "quota1",
						},
						models.Quota{
							GUID: "guid2",
							Name: "quota2",
						},
					)),
				),
			)
			objs, err := target.GetOrganizationQuotas()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(objs).Should(HaveLen(2))
			Ω(objs[0].GUID).Should(Equal("guid1"))
			Ω(objs[0].Name).Should(Equal("quota1"))
			Ω(objs[1].GUID).Should(Equal("guid2"))
			Ω(objs[1].Name).Should(Equal("quota2"))
		})
	})


	Context("fetching space quotas", func() {
		It("no error occurs", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v3/space_quotas", "per_page=5000"),
					ghttp.RespondWith(http.StatusOK, serializeList(
						models.Quota{
							GUID: "guid1",
							Name: "quota1",
						},
						models.Quota{
							GUID: "guid2",
							Name: "quota2",
						},
					)),
				),
			)
			objs, err := target.GetSpaceQuotas()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(objs).Should(HaveLen(2))
			Ω(objs[0].GUID).Should(Equal("guid1"))
			Ω(objs[0].Name).Should(Equal("quota1"))
			Ω(objs[1].GUID).Should(Equal("guid2"))
			Ω(objs[1].Name).Should(Equal("quota2"))
		})
	})


	Context("fetching space summary", func() {
		It("no error occurs", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/spaces/space-guid/summary"),
					ghttp.RespondWith(http.StatusOK, serialize(
						models.SpaceSummary{
							GUID: "space-guid",
							Apps: []models.AppSummary{
								models.AppSummary{
									GUID: "app1-guid",
									RunningInstances: 1,
								},
								models.AppSummary{
									GUID: "app2-guid",
									RunningInstances: 2,
								},
							},
						},
					)),
				),
			)
			objs, err := target.GetSpaceSummary("space-guid")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(objs.GUID).Should(Equal("space-guid"))
			Ω(objs.Apps).Should(HaveLen(2))
			Ω(objs.Apps[0].GUID).Should(Equal("app1-guid"))
			Ω(objs.Apps[0].RunningInstances).Should(Equal(1))
			Ω(objs.Apps[1].GUID).Should(Equal("app2-guid"))
			Ω(objs.Apps[1].RunningInstances).Should(Equal(2))
		})
	})

	Context("fetching app events", func() {
		It("no error occurs", func() {
			now := time.Now()
			before := now.Add(-10 * time.Minute)
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v3/audit_events"),
					ghttp.RespondWith(http.StatusOK, serializeList(
						models.Event{
							GUID: "event1-guid",
							CreatedAt: before,
							UpdatedAt: now,
							Type: "event1-type",
							Actor: models.EventActor{
								GUID: "event1-actor-guid",
								Type: "event1-actor-type",
								Name: "event1-actor-name",
							},
							Target: models.EventTarget{
								GUID: "event1-target-guid",
								Type: "event1-target-type",
								Name: "event1-target-name",
							},
							Space: models.EventSpace{
								GUID: "event1-space-guid",
							},
							Org: models.EventOrg{
								GUID: "event1-org-guid",
							},
						},
						models.Event{
							GUID: "event2-guid",
							CreatedAt: before,
							UpdatedAt: now,
							Type: "event2-type",
							Actor: models.EventActor{
								GUID: "event2-actor-guid",
								Type: "event2-actor-type",
								Name: "event2-actor-name",
							},
							Target: models.EventTarget{
								GUID: "event2-target-guid",
								Type: "event2-target-type",
								Name: "event2-target-name",
							},
							Space: models.EventSpace{
								GUID: "event2-space-guid",
							},
							Org: models.EventOrg{
								GUID: "event2-org-guid",
							},
						},
					)),
				),
			)
			objs, err := target.GetEvents()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(objs).Should(HaveLen(2))
			Ω(objs[0].GUID).Should(Equal("event1-guid"))
			Ω(objs[0].CreatedAt).Should(BeTemporally("==", before))
			Ω(objs[0].UpdatedAt).Should(BeTemporally("==", now))
			Ω(objs[0].Type).Should(Equal("event1-type"))
			Ω(objs[0].Actor.GUID).Should(Equal("event1-actor-guid"))
			Ω(objs[0].Actor.Type).Should(Equal("event1-actor-type"))
			Ω(objs[0].Actor.Name).Should(Equal("event1-actor-name"))
			Ω(objs[0].Target.GUID).Should(Equal("event1-target-guid"))
			Ω(objs[0].Target.Type).Should(Equal("event1-target-type"))
			Ω(objs[0].Target.Name).Should(Equal("event1-target-name"))
			Ω(objs[0].Space.GUID).Should(Equal("event1-space-guid"))
			Ω(objs[0].Org.GUID).Should(Equal("event1-org-guid"))
			Ω(objs[1].GUID).Should(Equal("event2-guid"))
			Ω(objs[1].CreatedAt).Should(BeTemporally("==", before))
			Ω(objs[1].UpdatedAt).Should(BeTemporally("==", now))
			Ω(objs[1].Type).Should(Equal("event2-type"))
			Ω(objs[1].Actor.GUID).Should(Equal("event2-actor-guid"))
			Ω(objs[1].Actor.Type).Should(Equal("event2-actor-type"))
			Ω(objs[1].Actor.Name).Should(Equal("event2-actor-name"))
			Ω(objs[1].Target.GUID).Should(Equal("event2-target-guid"))
			Ω(objs[1].Target.Type).Should(Equal("event2-target-type"))
			Ω(objs[1].Target.Name).Should(Equal("event2-target-name"))
			Ω(objs[1].Space.GUID).Should(Equal("event2-space-guid"))
			Ω(objs[1].Org.GUID).Should(Equal("event2-org-guid"))
		})
	})
})
