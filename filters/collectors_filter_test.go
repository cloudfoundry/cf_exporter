package filters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-community/cf_exporter/filters"
)

var _ = Describe("CollectorsFilter", func() {
	var (
		err     error
		filters []string

		collectorsFilter *CollectorsFilter
	)

	JustBeforeEach(func() {
		collectorsFilter, err = NewCollectorsFilter(filters)
	})

	Describe("New", func() {
		Context("when filters are supported", func() {
			BeforeEach(func() {
				filters = []string{
					ApplicationsCollector,
					ApplicationEventsCollector,
					OrganizationsCollector,
					SecurityGroupsCollector,
					ServiceInstancesCollector,
					ServicesCollector,
					SpacesCollector,
					StacksCollector,
				}
			})

			It("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when filters are not supported", func() {
			BeforeEach(func() {
				filters = []string{ApplicationsCollector, "Unknown"}
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Collector filter `Unknown` is not supported"))
			})
		})
	})

	Describe("Enabled", func() {
		BeforeEach(func() {
			filters = []string{ApplicationsCollector}
		})

		Context("when collector is enabled", func() {
			It("returns true", func() {
				Expect(collectorsFilter.Enabled(ApplicationsCollector)).To(BeTrue())
			})
		})

		Context("when collector is not enabled", func() {
			It("returns false", func() {
				Expect(collectorsFilter.Enabled(OrganizationsCollector)).To(BeFalse())
			})
		})

		Context("when there are no filters", func() {
			BeforeEach(func() {
				filters = []string{}
			})

			It("returns true", func() {
				Expect(collectorsFilter.Enabled(ApplicationsCollector)).To(BeTrue())
			})
		})
	})
})
