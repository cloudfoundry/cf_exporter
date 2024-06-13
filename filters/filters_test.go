package filters_test

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/cloudfoundry/cf_exporter/filters"
)

var _ = ginkgo.Describe("Filters", func() {

	ginkgo.Describe("Constructing Filters", func() {
		var (
			err error
			f   *filters.Filter
		)

		ginkgo.Context("with no active filters", func() {
			ginkgo.BeforeEach(func() {
				f, err = filters.NewFilter()
			})
			ginkgo.It("no error occurs", func() {
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("all but events are active", func() {
				gomega.Expect(f.Enabled(filters.Applications)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.Buildpacks)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.IsolationSegments)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.Organizations)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.Routes)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.SecurityGroups)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.ServiceBindings)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.ServicePlans)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.Services)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.Spaces)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.Stacks)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.Tasks)).To(gomega.BeFalse())
				gomega.Expect(f.Enabled(filters.Events)).To(gomega.BeFalse())
			})
		})

		ginkgo.Context("with valid active filters", func() {
			ginkgo.BeforeEach(func() {
				f, err = filters.NewFilter(filters.Applications, filters.Stacks)
			})
			ginkgo.It("no error occurs", func() {
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It("only given filters are active", func() {
				gomega.Expect(f.Enabled(filters.Applications)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.Buildpacks)).To(gomega.BeFalse())
				gomega.Expect(f.Enabled(filters.IsolationSegments)).To(gomega.BeFalse())
				gomega.Expect(f.Enabled(filters.Organizations)).To(gomega.BeFalse())
				gomega.Expect(f.Enabled(filters.Routes)).To(gomega.BeFalse())
				gomega.Expect(f.Enabled(filters.SecurityGroups)).To(gomega.BeFalse())
				gomega.Expect(f.Enabled(filters.ServiceBindings)).To(gomega.BeFalse())
				gomega.Expect(f.Enabled(filters.ServicePlans)).To(gomega.BeFalse())
				gomega.Expect(f.Enabled(filters.Services)).To(gomega.BeFalse())
				gomega.Expect(f.Enabled(filters.Spaces)).To(gomega.BeFalse())
				gomega.Expect(f.Enabled(filters.Stacks)).To(gomega.BeTrue())
				gomega.Expect(f.Enabled(filters.Tasks)).To(gomega.BeFalse())
				gomega.Expect(f.Enabled(filters.Events)).To(gomega.BeFalse())
			})

			ginkgo.It("querying all", func() {
				gomega.Expect(f.All(filters.Applications, filters.Stacks)).To(gomega.BeTrue())
				gomega.Expect(f.All(filters.Applications, filters.Spaces)).To(gomega.BeFalse())
				gomega.Expect(f.All()).To(gomega.BeTrue())
			})

			ginkgo.It("querying any", func() {
				gomega.Expect(f.Any(filters.Applications, filters.Stacks)).To(gomega.BeTrue())
				gomega.Expect(f.Any(filters.Applications, filters.Spaces)).To(gomega.BeTrue())
				gomega.Expect(f.Any(filters.Organizations, filters.Spaces)).To(gomega.BeFalse())
				gomega.Expect(f.Any()).To(gomega.BeFalse())
			})
		})

		ginkgo.Context("with invalid filters", func() {
			ginkgo.BeforeEach(func() {
				f, err = filters.NewFilter("I don't exist")
			})
			ginkgo.It("an error occurs", func() {
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

	})

})
