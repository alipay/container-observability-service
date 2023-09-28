package modules

var ShareModuleFactory = &ModuleFactory{
	generators: make(map[string]func() DeliveryModule, 10),
}

type ModuleFactory struct {
	generators map[string]func() DeliveryModule
}

func (m *ModuleFactory) Register(moduleName string, generator func() DeliveryModule) {
	m.generators[moduleName] = generator
}

func (m *ModuleFactory) GetModuleByName(moduleName string) DeliveryModule {
	generator := m.generators[moduleName]
	if generator != nil {
		return generator()
	}
	return nil
}
