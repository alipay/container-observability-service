package analyzers

var ShareAnalyzerFactory = &AnalyzerFactory{
	generators: make(map[DeliveryType]func() *DAGAnalyzer, 10),
}

type AnalyzerFactory struct {
	generators map[DeliveryType]func() *DAGAnalyzer
}

func (a *AnalyzerFactory) Register(deliveryType DeliveryType, generator func() *DAGAnalyzer) {
	a.generators[deliveryType] = generator
}

func (a *AnalyzerFactory) GetAnalyzerByType(deliveryType DeliveryType) *DAGAnalyzer {
	generator := a.generators[deliveryType]
	if generator != nil {
		return generator()
	}
	return nil
}
