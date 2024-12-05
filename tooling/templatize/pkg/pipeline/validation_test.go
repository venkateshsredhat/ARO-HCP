package pipeline

import (
	"testing"

	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
)

func TestRGValidate(t *testing.T) {
	testCases := []struct {
		name string
		rg   *ResourceGroup
		err  string
	}{
		{
			name: "missing name",
			rg:   &ResourceGroup{},
			err:  "resource group name is required",
		},
		{
			name: "missing subscription",
			rg:   &ResourceGroup{Name: "test"},
			err:  "subscription is required",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.rg.Validate()
			assert.Error(t, err, tc.err)
		})
	}

}

func TestPipelineValidate(t *testing.T) {
	testCases := []struct {
		name     string
		pipeline *Pipeline
		err      string
	}{
		{
			name: "missing name",
			pipeline: &Pipeline{
				ResourceGroups: []*ResourceGroup{{}},
			},
			err: "resource group name is required",
		},
		{
			name: "missing subscription",
			pipeline: &Pipeline{
				ResourceGroups: []*ResourceGroup{
					{
						Name: "rg",
					},
				},
			},
			err: "subscription is required",
		},
		{
			name: "missing step dependency",
			pipeline: &Pipeline{
				ResourceGroups: []*ResourceGroup{
					{
						Name:         "rg1",
						Subscription: "sub1",
						Steps: []*Step{
							{
								Name: "step1",
							},
						},
					},
					{
						Name:         "rg2",
						Subscription: "sub1",
						Steps: []*Step{
							{
								Name:      "step2",
								DependsOn: []string{"step3"},
							},
						},
					},
				},
			},
			err: "invalid dependency on step step2: dependency step3 does not exist",
		},
		{
			name: "duplicate step name",
			pipeline: &Pipeline{
				ResourceGroups: []*ResourceGroup{
					{
						Name:         "rg1",
						Subscription: "sub1",
						Steps: []*Step{
							{
								Name: "step1",
							},
						},
					},
					{
						Name:         "rg2",
						Subscription: "sub1",
						Steps: []*Step{
							{
								Name: "step1",
							},
						},
					},
				},
			},
			err: "duplicate step name \"step1\"",
		},
		{
			name: "valid step dependencies",
			pipeline: &Pipeline{
				ResourceGroups: []*ResourceGroup{
					{
						Name:         "rg1",
						Subscription: "sub1",
						Steps: []*Step{
							{
								Name: "step1",
							},
						},
					},
					{
						Name:         "rg2",
						Subscription: "sub1",
						Steps: []*Step{
							{
								Name:      "step2",
								DependsOn: []string{"step1"},
							},
						},
					},
				},
			},
			err: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.pipeline.Validate()
			if tc.err == "" {
				assert.NilError(t, err)
			} else {
				assert.Error(t, err, tc.err)
			}
		})
	}
}

func TestGetSchemaForPipeline(t *testing.T) {
	testCases := []struct {
		name              string
		pipeline          map[string]interface{}
		expectedSchemaRef string
		err               string
	}{
		{
			name:              "default schema",
			pipeline:          map[string]interface{}{},
			expectedSchemaRef: defaultSchemaRef,
		},
		{
			name: "explicit schema",
			pipeline: map[string]interface{}{
				"$schema": pipelineSchemaV1Ref,
			},
			expectedSchemaRef: pipelineSchemaV1Ref,
		},
		{
			name: "invalid schema",
			pipeline: map[string]interface{}{
				"$schema": "invalid",
			},
			expectedSchemaRef: "",
			err:               "unsupported schema reference: invalid",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, ref, err := getSchemaForPipeline(tc.pipeline)
			if tc.err == "" {
				assert.NilError(t, err)
				assert.Assert(t, schema != nil)
				if tc.expectedSchemaRef != "" {
					assert.Equal(t, ref, tc.expectedSchemaRef)
				}
			} else {
				assert.Error(t, err, tc.err)
			}
		})
	}
}

func TestValidatePipelineSchema(t *testing.T) {
	testCases := []struct {
		name              string
		pipeline          map[string]interface{}
		expectedSchemaRef string
		err               string
	}{
		{
			name: "valid shell",
			pipeline: map[string]interface{}{
				"serviceGroup": "test",
				"rolloutName":  "test",
				"resourceGroups": []interface{}{
					map[string]interface{}{
						"name":         "rg",
						"subscription": "sub",
						"aksCluster":   "aks",
						"steps": []interface{}{
							map[string]interface{}{
								"name":    "step",
								"action":  "Shell",
								"command": "echo hello",
							},
						},
					},
				},
			},
		},
		{
			name: "invalid",
			pipeline: map[string]interface{}{
				"serviceGroup": "test",
				"rolloutName":  "test",
				"resourceGroups": []interface{}{
					map[string]interface{}{
						"name":         "rg",
						"subscription": "sub",
						"aksCluster":   "aks",
						"steps": []interface{}{
							map[string]interface{}{
								"name":   "step",
								"action": "Shell",
							},
						},
					},
				},
			},
			err: "pipeline is not compliant with schema pipeline.schema.v1",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pipelineBytes, err := yaml.Marshal(tc.pipeline)
			assert.NilError(t, err)
			err = ValidatePipelineSchema(pipelineBytes)
			if tc.err == "" {
				assert.NilError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}
