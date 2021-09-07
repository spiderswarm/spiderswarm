package main

import (
	spsw "github.com/spiderswarm/spiderswarm/lib"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	workflow := &spsw.Workflow{
		Name:    "testWorkflow",
		Version: "v0.0.0.0.1",
		TaskTemplates: []spsw.TaskTemplate{
			spsw.TaskTemplate{
				TaskName: "GetHTML",
				Initial:  true,
				ActionTemplates: []spsw.ActionTemplate{
					spsw.ActionTemplate{
						Name:       "HTTP1",
						StructName: "HTTPAction",
						ConstructorParams: map[string]interface{}{
							"baseURL": "https://news.ycombinator.com/",
							"method":  "GET",
							"canFail": false,
						},
					},
					spsw.ActionTemplate{
						Name:              "UTF8Decode",
						StructName:        "UTF8DecodeAction",
						ConstructorParams: map[string]interface{}{},
					},
					spsw.ActionTemplate{
						Name:       "MakePromise",
						StructName: "TaskPromiseAction",
						ConstructorParams: map[string]interface{}{
							"inputNames": []string{"htmlStr1", "htmlStr2"},
							"taskName":   "ParseHTML",
						},
					},
				},
				DataPipeTemplates: []spsw.DataPipeTemplate{
					spsw.DataPipeTemplate{
						SourceActionName: "HTTP1",
						SourceOutputName: spsw.HTTPActionOutputBody,
						DestActionName:   "UTF8Decode",
						DestInputName:    spsw.UTF8DecodeActionInputBytes,
					},
					spsw.DataPipeTemplate{
						SourceActionName: "UTF8Decode",
						SourceOutputName: spsw.UTF8DecodeActionOutputStr,
						DestActionName:   "MakePromise",
						DestInputName:    "htmlStr1",
					},
					spsw.DataPipeTemplate{
						SourceActionName: "UTF8Decode",
						SourceOutputName: spsw.UTF8DecodeActionOutputStr,
						DestActionName:   "MakePromise",
						DestInputName:    "htmlStr2",
					},
					spsw.DataPipeTemplate{
						SourceActionName: "MakePromise",
						SourceOutputName: spsw.TaskPromiseActionOutputPromise,
						TaskOutputName:   "promise",
					},
				},
			},
			spsw.TaskTemplate{
				TaskName: "ParseHTML",
				Initial:  false,
				ActionTemplates: []spsw.ActionTemplate{
					spsw.ActionTemplate{
						Name:       "TitleExtraction",
						StructName: "XPathAction",
						ConstructorParams: map[string]interface{}{
							"xpath":      "//a[@class='storylink']/text()",
							"expectMany": true,
						},
					},
					spsw.ActionTemplate{
						Name:       "LinkExtraction",
						StructName: "XPathAction",
						ConstructorParams: map[string]interface{}{
							"xpath":      "//a[@class='storylink']/@href",
							"expectMany": true,
						},
					},
					spsw.ActionTemplate{
						Name:       "YieldItem",
						StructName: "FieldJoinAction",
						ConstructorParams: map[string]interface{}{
							"inputNames": []string{"title", "link"},
							"itemName":   "HNItem",
						},
					},
				},
				DataPipeTemplates: []spsw.DataPipeTemplate{
					spsw.DataPipeTemplate{
						TaskInputName:  "htmlStr1",
						DestActionName: "TitleExtraction",
						DestInputName:  spsw.XPathActionInputHTMLStr,
					},
					spsw.DataPipeTemplate{
						TaskInputName:  "htmlStr2",
						DestActionName: "LinkExtraction",
						DestInputName:  spsw.XPathActionInputHTMLStr,
					},
					spsw.DataPipeTemplate{
						SourceActionName: "TitleExtraction",
						SourceOutputName: spsw.XPathActionOutputStr,
						DestActionName:   "YieldItem",
						DestInputName:    "title",
					},
					spsw.DataPipeTemplate{
						SourceActionName: "LinkExtraction",
						SourceOutputName: spsw.XPathActionOutputStr,
						DestActionName:   "YieldItem",
						DestInputName:    "link",
					},
					spsw.DataPipeTemplate{
						SourceActionName: "YieldItem",
						SourceOutputName: spsw.FieldJoinActionOutputItem,
						TaskOutputName:   "items",
					},
				},
			},
		},
	}

	spew.Dump(workflow)

	/*
		items, err := workflow.Run()
		if err != nil {
			spew.Dump(err)
		} else {
			spew.Dump(items)
		}
	*/
}