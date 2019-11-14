package helper

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

var expectedResultDocument string = "# testStep\n\n\t## Description \n\nLong Test description\n\n\t\n\t## Prerequisites\n\t\n\tnone\n\n\t\n\t\n\t## Parameters\n\n| name | mandatory | default |\n| ---- | --------- | ------- |\n | param0 | No | val0 | \n  | param1 | No | <nil> | \n  | param2 | Yes | <nil> | \n ## Details\n * ` param0 ` :  param0 description \n  * ` param1 ` :  param1 description \n  * ` param2 ` :  param1 description \n \n\t\n\t## We recommend to define values of step parameters via [config.yml file](../configuration.md). \n\nIn following sections of the config.yml the configuration is possible:\n\n| parameter | general | step/stage |\n|-----------|---------|------------|\n | param0 | X |  | \n  | param1 |  |  | \n  | param2 |  |  | \n \n\t\n\t## Side effects\n\t\n\tnone\n\t\n\t## Exceptions\n\t\n\tnone\n\t\n\t## Example\n\n\tnone\n"

func configMetaDataMock(name string) (io.ReadCloser, error) {
	meta1 := `metadata:
  name: testStep
  description: Test description
  longDescription: |
    Long Test description
spec:
  inputs:
    params:
      - name: param0
        type: string
        description: param0 description
        default: val0
        scope:
        - GENERAL
        - PARAMETERS
        mandatory: true
      - name: param1
        type: string
        description: param1 description
        scope:
        - PARAMETERS
      - name: param2
        type: string
        description: param1 description
        scope:
        - PARAMETERS
        mandatory: true
`
	var r string
	switch name {
	case "test.yaml":
		r = meta1
	default:
		r = ""
	}
	return ioutil.NopCloser(strings.NewReader(r)), nil
}

func configOpenDocTemplateFileMock(docTemplateFilePath string) (io.ReadCloser, error) {
	meta1 := `# ${docGenStepName}

	## ${docGenDescription}
	
	## Prerequisites
	
	none

	## ${docJenkinsPluginDependencies}
	
	## ${docGenParameters}
	
	## ${docGenConfiguration}
	
	## Side effects
	
	none
	
	## Exceptions
	
	none
	
	## Example

	none
`
	switch docTemplateFilePath {
	case "testStep.md":
		return ioutil.NopCloser(strings.NewReader(meta1)), nil
	default:
		return ioutil.NopCloser(strings.NewReader("")), fmt.Errorf("Wrong Path: %v", docTemplateFilePath)
	}
}

var stepData config.StepData = config.StepData{
	Spec: config.StepSpec{
		Inputs: config.StepInputs{
			Parameters: []config.StepParameters{
				{Name: "param0", Scope: []string{"GENERAL"}, Type: "string", Default: "val0"},
			},
			Resources: []config.StepResources{
				{Name: "resource0", Type: "stash", Description: "val0"},
				{Name: "resource1", Type: "stash", Description: "val1"},
				{Name: "resource2", Type: "stash", Description: "val2"},
			},
		},
		Containers: []config.Container{
			{Name: "container0", Image: "image", WorkingDir: "workingdir", Shell: "shell",
				EnvVars: []config.EnvVar{
					{"envar.name0", "envar.value0"},
				},
			},
			{Name: "container1", Image: "image", WorkingDir: "workingdir",
				EnvVars: []config.EnvVar{
					{"envar.name1", "envar.value1"},
				},
			},
			{Name: "container2a", Command: []string{"command"}, ImagePullPolicy: "pullpolicy", Image: "image", WorkingDir: "workingdir",
				EnvVars: []config.EnvVar{
					{"envar.name2a", "envar.value2a"}},
				Conditions: []config.Condition{
					{Params: []config.Param{
						{"param.name2a", "param.value2a"},
					}},
				},
			},
			{Name: "container2b", Image: "image", WorkingDir: "workingdir",
				EnvVars: []config.EnvVar{
					{"envar.name2b", "envar.value2b"},
				},
				Conditions: []config.Condition{
					{Params: []config.Param{
						{"param.name2b", "param.value2b"},
					}},
				},
			},
		},
		Sidecars: []config.Container{
			{Name: "sidecar0", Command: []string{"command"}, ImagePullPolicy: "pullpolicy", Image: "image", WorkingDir: "workingdir", ReadyCommand: "readycommand",
				EnvVars: []config.EnvVar{
					{"envar.name3", "envar.value3"}},
				Conditions: []config.Condition{
					{Params: []config.Param{
						{"param.name0", "param.value0"},
					}},
				},
			},
		},
	},
}

var resultDocumentContent string

func docFileWriterMock(docTemplateFilePath string, data []byte, perm os.FileMode) error {

	resultDocumentContent = string(data)
	switch docTemplateFilePath {
	case "testStep.md":
		return nil
	default:
		return fmt.Errorf("Wrong Path: %v", docTemplateFilePath)
	}
}

func TestGenerateStepDocumentationSuccess(t *testing.T) {
	var stepData config.StepData
	contentMetaData, _ := configMetaDataMock("test.yaml")
	stepData.ReadPipelineStepData(contentMetaData)

	generateStepDocumentation(stepData, DocuHelperData{true, "", configOpenDocTemplateFileMock, docFileWriterMock})

	t.Run("Docu Generation Success", func(t *testing.T) {
		assert.Equal(t, expectedResultDocument, resultDocumentContent)
	})
}

func TestGenerateStepDocumentationError(t *testing.T) {
	var stepData config.StepData
	contentMetaData, _ := configMetaDataMock("test.yaml")
	stepData.ReadPipelineStepData(contentMetaData)

	err := generateStepDocumentation(stepData, DocuHelperData{true, "Dummy", configOpenDocTemplateFileMock, docFileWriterMock})

	t.Run("Docu Generation Success", func(t *testing.T) {
		assert.Error(t, err, fmt.Sprintf("Error occured: %v\n", err))
	})
}

func TestReadAndAdjustTemplate(t *testing.T) {

	t.Run("Success Case", func(t *testing.T) {

		tmpl, _ := configOpenDocTemplateFileMock("testStep.md")
		content := readAndAdjustTemplate(tmpl)

		cases := []struct {
			x, y string
		}{
			{"{{docGenStepName .}}", "${docGenStepName}"},
			{"{{docGenConfiguration .}}", "${docGenConfiguration}"},
			{"{{docGenParameters .}}", "${docGenParameters}"},
			{"{{docGenDescription .}}", "${docGenDescription}"},
			{"", "${docJenkinsPluginDependencies}"},
		}
		for _, c := range cases {
			if len(c.x) > 0 {
				assert.Contains(t, content, c.x)
			}
			if len(c.y) > 0 {
				assert.NotContains(t, content, c.y)
			}
		}
	})
}

func TestAddContainerContent(t *testing.T) {

	t.Run("Success Case", func(t *testing.T) {

		var m map[string]string = make(map[string]string)
		addContainerContent(&stepData, m)
		assert.Equal(t, 7, len(m))

		cases := []struct {
			x, want string
		}{
			{"containerCommand", "command"},
			{"containerShell", "shell"},
			{"dockerEnvVars", "envar.name0=envar.value0, envar.name1=envar.value1 <br>param.name2a=param.value2a:\\[envar.name2a=envar.value2a\\] <br>param.name2b=param.value2b:\\[envar.name2b=envar.value2b\\]"},
			{"dockerImage", "image, image <br>param.name2a=param.value2a:image <br>param.name2b=param.value2b:image"},
			{"dockerName", "container0, container1 <br>container2a <br>container2b <br>"},
			{"dockerPullImage", "true"},
			{"dockerWorkspace", "workingdir, workingdir <br>param.name2a=param.value2a:workingdir <br>param.name2b=param.value2b:workingdir"},
		}
		for _, c := range cases {
			assert.Contains(t, m, c.x)
			assert.True(t, len(m[c.x]) > 0)
			assert.True(t, strings.Contains(m[c.x], c.want), fmt.Sprintf("%v:%v", c.x, m[c.x]))
		}
	})
}
func TestAddSidecarContent(t *testing.T) {

	t.Run("Success Case", func(t *testing.T) {

		var m map[string]string = make(map[string]string)
		addSidecarContent(&stepData, m)
		assert.Equal(t, 7, len(m))

		cases := []struct {
			x, want string
		}{
			{"sidecarCommand", "command"},
			{"sidecarEnvVars", "envar.name3=envar.value3"},
			{"sidecarImage", "image"},
			{"sidecarName", "sidecar0"},
			{"sidecarPullImage", "true"},
			{"sidecarReadyCommand", "readycommand"},
			{"sidecarWorkspace", "workingdir"},
		}
		for _, c := range cases {
			assert.Contains(t, m, c.x)
			assert.True(t, len(m[c.x]) > 0)
			assert.Equal(t, c.want, m[c.x], fmt.Sprintf("%v:%v", c.x, m[c.x]))
		}
	})
}

func TestAddStashContent(t *testing.T) {

	t.Run("Success Case", func(t *testing.T) {

		var m map[string]string = make(map[string]string)
		addStashContent(&stepData, m)
		assert.Equal(t, 1, len(m))

		cases := []struct {
			x, want string
		}{
			{"stashContent", "resource0, resource1, resource2"},
		}
		for _, c := range cases {
			assert.Contains(t, m, c.x)
			assert.True(t, len(m[c.x]) > 0)
			assert.True(t, strings.Contains(m[c.x], c.want), fmt.Sprintf("%v:%v", c.x, m[c.x]))
		}
	})
}

func TestGetDocuContextDefaults(t *testing.T) {

	t.Run("Success Case", func(t *testing.T) {

		m := getDocuContextDefaults(&stepData)
		assert.Equal(t, 15, len(m))

		cases := []struct {
			x, want string
		}{
			{"stashContent", "resource0, resource1, resource2"},
			{"sidecarCommand", "command"},
			{"sidecarEnvVars", "envar.name3=envar.value3"},
			{"sidecarImage", "image"},
			{"sidecarName", "sidecar0"},
			{"sidecarPullImage", "true"},
			{"sidecarReadyCommand", "readycommand"},
			{"sidecarWorkspace", "workingdir"},
			{"containerCommand", "command"},
			{"containerShell", "shell"},
			{"dockerEnvVars", "envar.name0=envar.value0, envar.name1=envar.value1 <br>param.name2a=param.value2a:\\[envar.name2a=envar.value2a\\] <br>param.name2b=param.value2b:\\[envar.name2b=envar.value2b\\]"},
			{"dockerImage", "image, image <br>param.name2a=param.value2a:image <br>param.name2b=param.value2b:image"},
			{"dockerName", "container0, container1 <br>container2a <br>container2b <br>"},
			{"dockerPullImage", "true"},
			{"dockerWorkspace", "workingdir, workingdir <br>param.name2a=param.value2a:workingdir <br>param.name2b=param.value2b:workingdir"},
		}
		for _, c := range cases {
			assert.Contains(t, m, c.x)
			assert.True(t, len(m[c.x]) > 0)
			assert.True(t, strings.Contains(m[c.x], c.want), fmt.Sprintf("%v:%v", c.x, m[c.x]))
		}
	})
}