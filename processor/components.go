package main

import (
	"github.com/luke-segars/rufio/shared/structs"
)

func init() {
	components = append(components, ProcessedGameComponent{structs.Component{
		QueueName: "generate_processed_game",
		OutputCollection: "processed_games",
		OutputKey: "_id",
		Modules: []structs.IModule{
			Raw2ProcessedGameModule{structs.Module{
				InputCollection: "raw_games",
				// Need to convert field name to mongo equivalent.
				InputKey: "Metadata.DocumentId",
				MappingName: "game2rawgame",
			}},
			ProcessedLeague2ProcessedGameModule{structs.Module{
				InputCollection: "processed_leagues", 
				InputKey: "_id",
				MappingName: "game2summoner", 
			}},
		},
	}})
}

type ProcessedGameComponent struct {
	// Inherit from the base component class
	structs.Component
}

func (pgc ProcessedGameComponent) GetProcessedType() structs.Document {
	return structs.ProcessedGameDocument{}
}

/**
 * One component per OUTPUT type. If you want to generate three separate output collections,
 * create three components. Each component is responsible for handling tasks of a particular
 * type and contains MODULES that define how each task type is handled.
 */
 /*
type IComponent interface {
	GetProcessedType()	structs.Document
}

type Component struct {
	Modules 			[]Module
	QueueName			string
	OutputCollection	string
	// Field name for the primary index of OutputCollection.
	OutputKey			string
	mappings 			map[string]IDMap
}
*/