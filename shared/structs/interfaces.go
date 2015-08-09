package structs

import (
	"time"
)

type DocType int

const (
	EmptyDoc				DocType = iota
	RawGameDoc 				DocType = iota
	RawSummonerDoc			DocType = iota
	RawLeagueDoc			DocType = iota

	ProcessedGameDoc		DocType = iota
	ProcessedSummonerDoc	DocType = iota
	ProcessedLeagueDoc		DocType = iota
)

/**
 * DocType's are the parent of all document types. They are the most generic interface and can
 * be applied to both raw and processed docs.
 */
type Document interface {
	// Returns a unique ID for the document.
	GetId()				int
	GetType()			DocType
	GetCreationTime()	time.Time
}

/**
 * One component per OUTPUT type. If you want to generate three separate output collections,
 * create three components. Each component is responsible for handling tasks of a particular
 * type and contains MODULES that define how each task type is handled.
 */
type IComponent interface {
	GetProcessedType()		Document
	StoreMappings(map[string]IDMap)

	GetModules()			[]IModule
	GetQueueName()			string
	GetOutputCollection()	string
	GetOutputKey()			string
}

type Component struct {
	Modules 			[]IModule
	QueueName			string
	OutputCollection	string
	// Field name for the primary index of OutputCollection.
	OutputKey			string
	mappings 			map[string]IDMap
}

func (c Component) StoreMappings(mappings map[string]IDMap) {
	c.mappings = mappings
}


func (c Component) GetQueueName() string {
	return c.QueueName
}

func (c Component) GetOutputCollection() string {
	return c.OutputCollection
}

func (c Component) GetOutputKey() string {
	return c.OutputKey
}

func (c Component) GetModules() []IModule {
	return c.Modules
}

/**
 * One module per INPUT type that needs to be consumed to generate a given output. For example,
 * if a particular datatype (component) is built from data using two different sources, that component
 * should have two modules, one per input source. Each module has an additive effect on a given
 * processed object, so one module can set certain fields while a second module sets others and the
 * final object will have both groups of fields set.
 */
type IModule interface {
//	ValueCollection				string
//	ValueFieldName				string

//	mapping						IDMap
	GetMapping()						IDMap
	GetInputCollection()				string
	GetInputKey()						string
	// Join the documents for this module with the existing partial document.
	Join([]Document, Document, int64)	Document
}

type Module struct {
	InputCollection		string
	// Field name for the primary index of InputCollection.
	InputKey			string

	mapping 			map[string]IDMap
	MappingName			string
}

func (m Module) GetMapping() IDMap {
	return m.mapping[m.MappingName]
}

func (m Module) GetInputCollection() string {
	return m.InputCollection
}

func (m Module) GetInputKey() string {
	return m.InputKey
}