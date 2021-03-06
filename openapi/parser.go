package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/grahambrooks/scheme/search"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"strings"
)

type Parser struct {
}

func (parser *Parser) ParseJson(reader io.Reader) (search.Model, error) {
	buffer, err := ioutil.ReadAll(reader)
	decoder := json.NewDecoder(bytes.NewReader(buffer))
	var spec map[string]interface{}
	err = decoder.Decode(&spec)
	if err != nil {
		return search.Model{}, fmt.Errorf("failed decoding OpenAPI spec %v", err)
	}
	model := search.Model{}
	switch spec["swagger"] {
	case "2.0":
		model.Kind = search.OpenAPI2
		info, exists := spec["info"]
		if exists {
			model.Title = info.(map[string]interface{})["title"].(string)
			description, exists := info.(map[string]interface{})["description"]
			if exists {
				model.Description = description.(string)
			}
			model.Version = info.(map[string]interface{})["version"].(string)
		}

		model.Resources = parseResources(spec)
	default:
		return search.Model{}, fmt.Errorf("unrecognized API version %s", spec["swagger"])
	}
	return model, nil
}

type SpecInfo struct {
	Description string `json:"description"`
	Version     string `json:"version"`
	Title       string `json:"title"`
}

type OpenAPISpec struct {
	OpenAPI string                 `json:"openapi"`
	Swagger string                 `json:"swagger"`
	Info    SpecInfo               `json:"info"`
	Paths   map[string]interface{} `json:"paths"`
}

func (parser *Parser) ParseRawYaml(reader io.Reader) (interface{}, error) {
	buffer, err := ioutil.ReadAll(reader)
	decoder := yaml.NewDecoder(bytes.NewReader(buffer))
	var spec map[string]interface{}
	err = decoder.Decode(&spec)

	return spec, err
}

func (parser *Parser) ParseRawJson(reader io.Reader) (interface{}, error) {
	buffer, err := ioutil.ReadAll(reader)
	decoder := json.NewDecoder(bytes.NewReader(buffer))
	var spec interface{}
	err = decoder.Decode(&spec)

	return spec, err
}

func (parser *Parser) ParseYaml(reader io.Reader) (search.Model, error) {
	buffer, err := ioutil.ReadAll(reader)
	decoder := yaml.NewDecoder(bytes.NewReader(buffer))
	var spec OpenAPISpec
	err = decoder.Decode(&spec)
	if err != nil {
		return search.Model{}, fmt.Errorf("failed decoding OpenAPI spec %v", err)
	}
	model := search.Model{}

	switch {
	case spec.Swagger == "2.0":
		model.Kind = search.OpenAPI2
		model.Title = spec.Info.Title
		model.Description = spec.Info.Description
		model.Version = spec.Info.Version

		for key := range spec.Paths {
			model.Resources = append(model.Resources, search.Resource{Path: key})
		}
	case strings.HasPrefix(spec.OpenAPI, "3"):
		model.Kind = search.OpenAPI3
		model.Title = spec.Info.Title
		model.Description = spec.Info.Description
		model.Version = spec.Info.Version

		for key := range spec.Paths {
			model.Resources = append(model.Resources, search.Resource{Path: key})
		}

	default:
		return search.Model{}, fmt.Errorf("unrecognized API version %s", strings.Join([]string{spec.Swagger, spec.OpenAPI}, "/"))
	}
	return model, nil
}

func parseResources(spec map[string]interface{}) []search.Resource {
	var resources []search.Resource

	if spec["paths"] != nil {
		paths := spec["paths"].(map[string]interface{})

		for key := range paths {
			resources = append(resources, search.Resource{Path: key})
		}
	}

	return resources
}
