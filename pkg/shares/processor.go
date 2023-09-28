package shares

import (
	"k8s.io/klog/v2"
)

var (
	BaseObjectProcessor = NewBaseProcessor()
	MilestoneProcessor  = NewMilestoneExtractor()

	metaProcessor Processor = &MetaProcessor{
		objectProcessor:    BaseObjectProcessor,
		milestoneExtractor: MilestoneProcessor,
	}
)

type Processor interface {
	CanProcess(event *AuditEvent) bool
	Process(event *AuditEvent) error
}

// BaseProcessor 用于基础的信息提取
type BaseProcessor struct {
	processors map[string]Processor
}

func NewBaseProcessor() *BaseProcessor {
	return &BaseProcessor{
		processors: make(map[string]Processor),
	}
}

func (l *BaseProcessor) Process(event *AuditEvent) error {
	for name, p := range l.processors {
		if p.CanProcess(event) {
			if err := p.Process(event); err != nil {
				klog.Errorf("processor %s is error: %v", name, err)
				return err
			}
		}
	}
	return nil
}

func (l *BaseProcessor) CanProcess(event *AuditEvent) bool {
	return true
}

func (l *BaseProcessor) Register(name string, p Processor) {
	l.processors[name] = p
}

// MilestoneExtractor 用于 hyper event 生命周期提取
type MilestoneExtractor struct {
	extractors map[string]Processor
}

func NewMilestoneExtractor() *MilestoneExtractor {
	return &MilestoneExtractor{
		extractors: make(map[string]Processor),
	}
}

func (l *MilestoneExtractor) Process(event *AuditEvent) error {
	for name, p := range l.extractors {
		if p.CanProcess(event) {
			if err := p.Process(event); err != nil {
				klog.Errorf("extractors %s is error: %v", name, err)
				return err
			}
		}
	}
	return nil
}

func (l *MilestoneExtractor) CanProcess(event *AuditEvent) bool {
	return true
}

func (l *MilestoneExtractor) Register(name string, p Processor) {
	l.extractors[name] = p
}

// MetaProcessor 元数据处理总入口
type MetaProcessor struct {
	objectProcessor    Processor
	milestoneExtractor Processor
}

func (m *MetaProcessor) CanProcess(event *AuditEvent) bool {
	return true
}

func (m *MetaProcessor) Process(event *AuditEvent) error {
	if err := m.objectProcessor.Process(event); err != nil {
		return err
	}

	if err := m.milestoneExtractor.Process(event); err != nil {
		return err
	}

	return nil
}
