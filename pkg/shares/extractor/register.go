package extractor

import "github.com/alipay/container-observability-service/pkg/shares"

func init() {
	shares.MilestoneProcessor.Register("PodCreateProcessor", &PodCreateProcessor{})
	shares.MilestoneProcessor.Register("PodDeleteProcessor", &PodDeleteProcessor{})
	shares.MilestoneProcessor.Register("PodPatchProcessor", &PodPatchProcessor{})
	shares.MilestoneProcessor.Register("PodUpdateProcessor", &PodUpdateProcessor{})
	shares.MilestoneProcessor.Register("PodBindingProcessor", &PodBindingProcessor{})
	shares.MilestoneProcessor.Register("PodEventProcessor", &PodEventProcessor{})
}
