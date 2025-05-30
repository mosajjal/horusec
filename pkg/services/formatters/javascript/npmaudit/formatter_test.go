// Copyright 2020 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package npmaudit

import (
	"errors"
	"testing"

	"github.com/ZupIT/horusec-devkit/pkg/entities/analysis"
	"github.com/ZupIT/horusec-devkit/pkg/enums/languages"
	"github.com/ZupIT/horusec-devkit/pkg/enums/tools"
	"github.com/stretchr/testify/assert"

	"github.com/mosajjal/horusec/config"
	"github.com/mosajjal/horusec/pkg/entities/toolsconfig"
	"github.com/mosajjal/horusec/pkg/services/formatters"
	"github.com/mosajjal/horusec/pkg/utils/testutil"
)

func TestNpmAuditParseOutput(t *testing.T) {
	t.Run("should add 1 vulnerabilities on analysis with no errors", func(t *testing.T) {
		analysis := new(analysis.Analysis)

		dockerAPIControllerMock := testutil.NewDockerMock()
		dockerAPIControllerMock.On("SetAnalysisID")
		dockerAPIControllerMock.On("CreateLanguageAnalysisContainer").Return(output, nil)

		service := formatters.NewFormatterService(analysis, dockerAPIControllerMock, newTestConfig(t, analysis))
		formatter := NewFormatter(service)
		formatter.StartAnalysis("")

		assert.Equal(t, 1, len(analysis.AnalysisVulnerabilities))
		for _, v := range analysis.AnalysisVulnerabilities {
			vuln := v.Vulnerability

			assert.Equal(t, tools.NpmAudit, vuln.SecurityTool)
			assert.Equal(t, languages.Javascript, vuln.Language)
			assert.NotEmpty(t, vuln.Details, "Expected not empty details")
			assert.NotEmpty(t, vuln.Code, "Expected not empty code")
			assert.NotEmpty(t, vuln.File, "Expected not empty file name")
			assert.NotEmpty(t, vuln.Line, "Expected not empty line")
			assert.NotEmpty(t, vuln.Severity, "Expected not empty severity")
		}
	})
	t.Run("Should parse output empty with no errors", func(t *testing.T) {
		analysis := new(analysis.Analysis)

		dockerAPIControllerMock := testutil.NewDockerMock()
		dockerAPIControllerMock.On("SetAnalysisID")
		dockerAPIControllerMock.On("CreateLanguageAnalysisContainer").Return("", nil)

		service := formatters.NewFormatterService(analysis, dockerAPIControllerMock, newTestConfig(t, analysis))
		formatter := NewFormatter(service)
		formatter.StartAnalysis("")

		assert.Equal(t, 0, len(analysis.AnalysisVulnerabilities))
		assert.False(t, analysis.HasErrors(), "Expected no errors on analysis")
	})

	t.Run("Should add error on analysis when parse output with not found error", func(t *testing.T) {
		analysis := new(analysis.Analysis)

		dockerAPIControllerMock := testutil.NewDockerMock()
		dockerAPIControllerMock.On("SetAnalysisID")
		dockerAPIControllerMock.On("CreateLanguageAnalysisContainer").Return("ERROR_PACKAGE_LOCK_NOT_FOUND", nil)

		service := formatters.NewFormatterService(analysis, dockerAPIControllerMock, newTestConfig(t, analysis))
		formatter := NewFormatter(service)
		formatter.StartAnalysis("")

		assert.Equal(t, 0, len(analysis.AnalysisVulnerabilities))
		assert.True(t, analysis.HasErrors(), "Expected errors on analysis")
	})

	t.Run("Should add error on analysis when parse invalid output", func(t *testing.T) {
		analysis := new(analysis.Analysis)

		dockerAPIControllerMock := testutil.NewDockerMock()
		dockerAPIControllerMock.On("SetAnalysisID")
		dockerAPIControllerMock.On("CreateLanguageAnalysisContainer").Return("invalid", nil)

		service := formatters.NewFormatterService(analysis, dockerAPIControllerMock, newTestConfig(t, analysis))
		formatter := NewFormatter(service)
		formatter.StartAnalysis("")

		assert.True(t, analysis.HasErrors(), "Expected no errors on analysis")
	})

	t.Run("should add error of executing container on analysis", func(t *testing.T) {
		analysis := new(analysis.Analysis)

		dockerAPIControllerMock := testutil.NewDockerMock()
		dockerAPIControllerMock.On("SetAnalysisID")
		dockerAPIControllerMock.On("CreateLanguageAnalysisContainer").Return("", errors.New("test"))

		service := formatters.NewFormatterService(analysis, dockerAPIControllerMock, newTestConfig(t, analysis))
		formatter := NewFormatter(service)
		formatter.StartAnalysis("")

		assert.True(t, analysis.HasErrors(), "Expected errors on analysis")
	})
	t.Run("Should not execute tool because it's ignored", func(t *testing.T) {
		analysis := new(analysis.Analysis)

		dockerAPIControllerMock := testutil.NewDockerMock()

		config := config.New()
		config.ToolsConfig = toolsconfig.ToolsConfig{
			tools.NpmAudit: toolsconfig.Config{
				IsToIgnore: true,
			},
		}
		service := formatters.NewFormatterService(analysis, dockerAPIControllerMock, config)
		formatter := NewFormatter(service)
		formatter.StartAnalysis("")
	})
}

func newTestConfig(t *testing.T, analysiss *analysis.Analysis) *config.Config {
	cfg := config.New()
	cfg.ProjectPath = testutil.CreateHorusecAnalysisDirectory(t, analysiss, testutil.JavaScriptExample1)
	return cfg
}

const output = `
{
  "advisories": {
    "1469": {
      "findings": [
        {
          "version": "0.6.6",
          "paths": [
            "express>qs"
          ]
        }
      ],
      "id": 1469,
      "created": "2020-02-10T19:09:50.604Z",
      "updated": "2020-02-14T22:24:16.925Z",
      "deleted": null,
      "title": "Prototype Pollution Protection Bypass",
      "found_by": {
        "link": "",
        "name": "Unknown",
        "email": ""
      },
      "reported_by": {
        "link": "",
        "name": "Unknown",
        "email": ""
      },
      "module_name": "qs",
      "cves": [
        "CVE-2017-1000048"
      ],
      "vulnerable_versions": "<6.0.4 || >=6.1.0 <6.1.2 || >=6.2.0 <6.2.3 || >=6.3.0 <6.3.2",
      "patched_versions": ">=6.0.4 <6.1.0 || >=6.1.2 <6.2.0 || >=6.2.3 <6.3.0 || >=6.3.2",
      "overview": "Affected version of qs are vulnerable to Prototype Pollution because it is possible to bypass the protection. The qs.parse function fails to properly prevent an object's prototype to be altered when parsing arbitrary input. Input containing [ or ] may bypass the prototype pollution protection and alter the Object prototype. This allows attackers to override properties that will exist in all objects, which may lead to Denial of Service or Remote Code Execution in specific circumstances.",
      "recommendation": "Upgrade to 6.0.4, 6.1.2, 6.2.3, 6.3.2 or later.",
      "references": "- [GitHub Issue](https://github.com/ljharb/qs/issues/200)\n- [Snyk Report](https://snyk.io/vuln/npm:qs:20170213)",
      "access": "public",
      "severity": "high",
      "cwe": "CWE-471",
      "metadata": {
        "module_type": "",
        "exploitability": 4,
        "affected_components": ""
      },
      "url": "https://npmjs.com/advisories/1469"
    }
  },
  "metadata": {
    "vulnerabilities": {
      "info": 0,
      "low": 8,
      "moderate": 6,
      "high": 7,
      "critical": 0
    },
    "dependencies": 23,
    "devDependencies": 0,
    "optionalDependencies": 0,
    "totalDependencies": 23
  },
  "runId": "7c3c5266-3f9d-4924-a8b7-93fad66e64e0"
}
`
