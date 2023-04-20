package filters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bosh-prometheus/cf_exporter/filters"
)

var _ = Describe("Filters", func() {

	Describe("Constructing Filters", func() {
		var (
			err error
			f   *filters.Filter
		)

		Context("with no active filters", func() {
			BeforeEach(func() {
				f, err = filters.NewFilter()
			})
			It("no error occurs", func() {
				Expect(err).To(BeNil())
			})
			It("all but events are active", func() {
				Expect(f.Enabled(filters.Applications)).To(BeTrue())
				Expect(f.Enabled(filters.Buildpacks)).To(BeTrue())
				Expect(f.Enabled(filters.IsolationSegments)).To(BeTrue())
				Expect(f.Enabled(filters.Organizations)).To(BeTrue())
				Expect(f.Enabled(filters.Routes)).To(BeTrue())
				Expect(f.Enabled(filters.SecurityGroups)).To(BeTrue())
				Expect(f.Enabled(filters.ServiceBindings)).To(BeTrue())
				Expect(f.Enabled(filters.ServicePlans)).To(BeTrue())
				Expect(f.Enabled(filters.Services)).To(BeTrue())
				Expect(f.Enabled(filters.Spaces)).To(BeTrue())
				Expect(f.Enabled(filters.Stacks)).To(BeTrue())
				Expect(f.Enabled(filters.Events)).To(BeFalse())
			})
		})

		Context("with valid active filters", func() {
			BeforeEach(func() {
				f, err = filters.NewFilter(filters.Applications, filters.Stacks)
			})
			It("no error occurs", func() {
				Expect(err).To(BeNil())
			})
			It("only given filters are active", func() {
				Expect(f.Enabled(filters.Applications)).To(BeTrue())
				Expect(f.Enabled(filters.Buildpacks)).To(BeFalse())
				Expect(f.Enabled(filters.IsolationSegments)).To(BeFalse())
				Expect(f.Enabled(filters.Organizations)).To(BeFalse())
				Expect(f.Enabled(filters.Routes)).To(BeFalse())
				Expect(f.Enabled(filters.SecurityGroups)).To(BeFalse())
				Expect(f.Enabled(filters.ServiceBindings)).To(BeFalse())
				Expect(f.Enabled(filters.ServicePlans)).To(BeFalse())
				Expect(f.Enabled(filters.Services)).To(BeFalse())
				Expect(f.Enabled(filters.Spaces)).To(BeFalse())
				Expect(f.Enabled(filters.Stacks)).To(BeTrue())
				Expect(f.Enabled(filters.Events)).To(BeFalse())
			})

			It("querying all", func() {
				Expect(f.All(filters.Applications, filters.Stacks)).To(BeTrue())
				Expect(f.All(filters.Applications, filters.Spaces)).To(BeFalse())
				Expect(f.All()).To(BeTrue())
			})

			It("querying any", func() {
				Expect(f.Any(filters.Applications, filters.Stacks)).To(BeTrue())
				Expect(f.Any(filters.Applications, filters.Spaces)).To(BeTrue())
				Expect(f.Any(filters.Organizations, filters.Spaces)).To(BeFalse())
				Expect(f.Any()).To(BeFalse())
			})
		})

		Context("with invalid filters", func() {
			BeforeEach(func() {
				f, err = filters.NewFilter("I don't exist")
			})
			It("an error occurs", func() {
				Expect(err).NotTo(BeNil())
			})
		})

	})

})
