package filters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/bosh-prometheus/cf_exporter/filters"
)

var _ = Describe("CollectorsFilter", func() {
	var (
		err     error
		filters []string

		collectorsFilter *CollectorsFilter
		cfAPIv3Enabled   bool
	)

	JustBeforeEach(func() {
		collectorsFilter, err = NewCollectorsFilter(filters, cfAPIv3Enabled)
	})

	Describe("New", func() {
		Context("when filters are supported", func() {
			BeforeEach(func() {
				filters = []string{
					ApplicationsCollector,
					OrganizationsCollector,
					RoutesCollector,
					SecurityGroupsCollector,
					ServiceBindingsCollector,
					ServiceInstancesCollector,
					ServicePlansCollector,
					ServicesCollector,
					SpacesCollector,
					StacksCollector,
					EventsCollector,
				}
			})

			It("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			Context("and has leading and/or trailing whitespaces", func() {
				BeforeEach(func() {
					filters = []string{"   " + ApplicationsCollector + "  "}
				})

				It("does not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("when CF API V3 is required", func() {
				BeforeEach(func() {
					filters = append(filters, IsolationSegmentsCollector)
				})

				Context("and is enabled", func() {
					BeforeEach(func() {
						cfAPIv3Enabled = true
					})

					It("does not return an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("and is not enabled", func() {
					BeforeEach(func() {
						cfAPIv3Enabled = false
					})

					It("returns an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("IsolationSegments Collector filter need CF API V3 enabled"))
					})
				})
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

			Context("when CF API V3 is enabled", func() {
				BeforeEach(func() {
					cfAPIv3Enabled = true
				})

				It("Applications Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(ApplicationsCollector)).To(BeTrue())
				})

				It("Isolation Segments Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(IsolationSegmentsCollector)).To(BeTrue())
				})

				It("Organizations Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(OrganizationsCollector)).To(BeTrue())
				})

				It("Routes Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(RoutesCollector)).To(BeTrue())
				})

				It("Security Groups Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(SecurityGroupsCollector)).To(BeTrue())
				})

				It("Service Bindings Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(ServiceBindingsCollector)).To(BeTrue())
				})

				It("Service Instances Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(ServiceInstancesCollector)).To(BeTrue())
				})

				It("Service Plans Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(ServicePlansCollector)).To(BeTrue())
				})

				It("Services Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(ServicesCollector)).To(BeTrue())
				})

				It("Spaces Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(SpacesCollector)).To(BeTrue())
				})

				It("Stacks Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(StacksCollector)).To(BeTrue())
				})
			})

			Context("when CF API V3 is not enabled", func() {
				BeforeEach(func() {
					cfAPIv3Enabled = false
				})

				It("Applications Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(ApplicationsCollector)).To(BeTrue())
				})

				It("Isolation Segments Collector should be disabled", func() {
					Expect(collectorsFilter.Enabled(IsolationSegmentsCollector)).To(BeFalse())
				})

				It("Organizations Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(OrganizationsCollector)).To(BeTrue())
				})

				It("Routes Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(RoutesCollector)).To(BeTrue())
				})

				It("Security Groups Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(SecurityGroupsCollector)).To(BeTrue())
				})

				It("Service Bindings Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(ServiceBindingsCollector)).To(BeTrue())
				})

				It("Service Instances Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(ServiceInstancesCollector)).To(BeTrue())
				})

				It("Service Plans Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(ServicePlansCollector)).To(BeTrue())
				})

				It("Services Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(ServicesCollector)).To(BeTrue())
				})

				It("Spaces Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(SpacesCollector)).To(BeTrue())
				})

				It("Stacks Collector should be enabled", func() {
					Expect(collectorsFilter.Enabled(StacksCollector)).To(BeTrue())
				})
			})

			It("returns false", func() {
				Expect(collectorsFilter.Enabled(EventsCollector)).To(BeFalse())
			})
		})
	})
})
