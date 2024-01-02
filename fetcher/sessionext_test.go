package fetcher

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"

	"github.com/onsi/gomega/ghttp"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"encoding/json"
	"net/http"

	"github.com/bosh-prometheus/cf_exporter/models"
)

const (
	//nolint:gosec
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
	gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
	return string(content)
}

var _ = ginkgo.Describe("Extensions", func() {

	var (
		err    error
		server *ghttp.Server
		target *SessionExt
		config *CFConfig
	)

	ginkgo.BeforeEach(func() {
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
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		gomega.Ω(target).ShouldNot(gomega.BeNil())
	})

	ginkgo.AfterEach(func() {
		server.Close()
	})

	ginkgo.Context("fetching applications", func() {
		ginkgo.It("no error occurs", func() {
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
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
			gomega.Ω(app).Should(gomega.HaveLen(1))
			gomega.Ω(app[0].GUID).Should(gomega.Equal("app1-guid"))
			gomega.Ω(app[0].Name).Should(gomega.Equal("app1"))
		})
	})

	ginkgo.Context("fetching tasks", func() {
		ginkgo.It("no error occurs", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v3/tasks", "per_page=5000&states=PENDING,RUNNING,CANCELING"),
					ghttp.RespondWith(http.StatusOK, serializeList(
						models.Task{
							GUID:  "guid1",
							State: constant.TaskPending,
						},
						models.Task{
							GUID:  "guid2",
							State: constant.TaskCanceling,
						},
					)),
				),
			)
			objs, err := target.GetTasks()
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
			gomega.Ω(objs).Should(gomega.HaveLen(2))
			gomega.Ω(objs[0].GUID).Should(gomega.Equal("guid1"))
			gomega.Ω(objs[0].State).Should(gomega.Equal(constant.TaskPending))
			gomega.Ω(objs[1].GUID).Should(gomega.Equal("guid2"))
			gomega.Ω(objs[1].State).Should(gomega.Equal(constant.TaskCanceling))
		})
	})

	ginkgo.Context("fetching org quotas", func() {
		ginkgo.It("no error occurs", func() {
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
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
			gomega.Ω(objs).Should(gomega.HaveLen(2))
			gomega.Ω(objs[0].GUID).Should(gomega.Equal("guid1"))
			gomega.Ω(objs[0].Name).Should(gomega.Equal("quota1"))
			gomega.Ω(objs[1].GUID).Should(gomega.Equal("guid2"))
			gomega.Ω(objs[1].Name).Should(gomega.Equal("quota2"))
		})
	})

	ginkgo.Context("fetching space quotas", func() {
		ginkgo.It("no error occurs", func() {
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
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
			gomega.Ω(objs).Should(gomega.HaveLen(2))
			gomega.Ω(objs[0].GUID).Should(gomega.Equal("guid1"))
			gomega.Ω(objs[0].Name).Should(gomega.Equal("quota1"))
			gomega.Ω(objs[1].GUID).Should(gomega.Equal("guid2"))
			gomega.Ω(objs[1].Name).Should(gomega.Equal("quota2"))
		})
	})

	ginkgo.Context("fetching space summary", func() {
		ginkgo.It("no error occurs", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/spaces/space-guid/summary"),
					ghttp.RespondWith(http.StatusOK, serialize(
						models.SpaceSummary{
							GUID: "space-guid",
							Apps: []models.AppSummary{
								{
									GUID:             "app1-guid",
									RunningInstances: 1,
								},
								{
									GUID:             "app2-guid",
									RunningInstances: 2,
								},
							},
						},
					)),
				),
			)
			objs, err := target.GetSpaceSummary("space-guid")
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
			gomega.Ω(objs.GUID).Should(gomega.Equal("space-guid"))
			gomega.Ω(objs.Apps).Should(gomega.HaveLen(2))
			gomega.Ω(objs.Apps[0].GUID).Should(gomega.Equal("app1-guid"))
			gomega.Ω(objs.Apps[0].RunningInstances).Should(gomega.Equal(1))
			gomega.Ω(objs.Apps[1].GUID).Should(gomega.Equal("app2-guid"))
			gomega.Ω(objs.Apps[1].RunningInstances).Should(gomega.Equal(2))
		})
	})

	ginkgo.Context("fetching app events", func() {
		ginkgo.It("no error occurs", func() {
			now := time.Now()
			before := now.Add(-10 * time.Minute)
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v3/audit_events"),
					ghttp.RespondWith(http.StatusOK, serializeList(
						models.Event{
							GUID:      "event1-guid",
							CreatedAt: before,
							UpdatedAt: now,
							Type:      "event1-type",
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
							GUID:      "event2-guid",
							CreatedAt: before,
							UpdatedAt: now,
							Type:      "event2-type",
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
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
			gomega.Ω(objs).Should(gomega.HaveLen(2))
			gomega.Ω(objs[0].GUID).Should(gomega.Equal("event1-guid"))
			gomega.Ω(objs[0].CreatedAt).Should(gomega.BeTemporally("==", before))
			gomega.Ω(objs[0].UpdatedAt).Should(gomega.BeTemporally("==", now))
			gomega.Ω(objs[0].Type).Should(gomega.Equal("event1-type"))
			gomega.Ω(objs[0].Actor.GUID).Should(gomega.Equal("event1-actor-guid"))
			gomega.Ω(objs[0].Actor.Type).Should(gomega.Equal("event1-actor-type"))
			gomega.Ω(objs[0].Actor.Name).Should(gomega.Equal("event1-actor-name"))
			gomega.Ω(objs[0].Target.GUID).Should(gomega.Equal("event1-target-guid"))
			gomega.Ω(objs[0].Target.Type).Should(gomega.Equal("event1-target-type"))
			gomega.Ω(objs[0].Target.Name).Should(gomega.Equal("event1-target-name"))
			gomega.Ω(objs[0].Space.GUID).Should(gomega.Equal("event1-space-guid"))
			gomega.Ω(objs[0].Org.GUID).Should(gomega.Equal("event1-org-guid"))
			gomega.Ω(objs[1].GUID).Should(gomega.Equal("event2-guid"))
			gomega.Ω(objs[1].CreatedAt).Should(gomega.BeTemporally("==", before))
			gomega.Ω(objs[1].UpdatedAt).Should(gomega.BeTemporally("==", now))
			gomega.Ω(objs[1].Type).Should(gomega.Equal("event2-type"))
			gomega.Ω(objs[1].Actor.GUID).Should(gomega.Equal("event2-actor-guid"))
			gomega.Ω(objs[1].Actor.Type).Should(gomega.Equal("event2-actor-type"))
			gomega.Ω(objs[1].Actor.Name).Should(gomega.Equal("event2-actor-name"))
			gomega.Ω(objs[1].Target.GUID).Should(gomega.Equal("event2-target-guid"))
			gomega.Ω(objs[1].Target.Type).Should(gomega.Equal("event2-target-type"))
			gomega.Ω(objs[1].Target.Name).Should(gomega.Equal("event2-target-name"))
			gomega.Ω(objs[1].Space.GUID).Should(gomega.Equal("event2-space-guid"))
			gomega.Ω(objs[1].Org.GUID).Should(gomega.Equal("event2-org-guid"))
		})
	})
})
