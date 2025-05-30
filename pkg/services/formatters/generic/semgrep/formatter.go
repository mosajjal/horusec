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

package semgrep

import (
	"encoding/json"
	"path/filepath"
	"strconv"

	"github.com/ZupIT/horusec-devkit/pkg/entities/vulnerability"
	"github.com/ZupIT/horusec-devkit/pkg/enums/languages"
	"github.com/ZupIT/horusec-devkit/pkg/enums/severities"
	"github.com/ZupIT/horusec-devkit/pkg/enums/tools"
	"github.com/ZupIT/horusec-devkit/pkg/utils/logger"

	"github.com/mosajjal/horusec/pkg/entities/docker"
	"github.com/mosajjal/horusec/pkg/enums/images"
	"github.com/mosajjal/horusec/pkg/helpers/messages"
	"github.com/mosajjal/horusec/pkg/services/formatters"
	vulnhash "github.com/mosajjal/horusec/pkg/utils/vuln_hash"
)

type Formatter struct {
	formatters.IService
}

func NewFormatter(service formatters.IService) formatters.IFormatter {
	return &Formatter{
		service,
	}
}

func (f *Formatter) StartAnalysis(projectSubPath string) {
	if f.ToolIsToIgnore(tools.Semgrep) || f.IsDockerDisabled() {
		logger.LogDebugWithLevel(messages.MsgDebugToolIgnored + tools.Semgrep.ToString())
		return
	}

	output, err := f.startSemgrep(projectSubPath)
	f.SetAnalysisError(err, tools.Semgrep, output, projectSubPath)
	f.LogDebugWithReplace(messages.MsgDebugToolFinishAnalysis, tools.Semgrep, languages.Generic)
}

func (f *Formatter) startSemgrep(projectSubPath string) (string, error) {
	f.LogDebugWithReplace(messages.MsgDebugToolStartAnalysis, tools.Semgrep, languages.Generic)

	output, err := f.ExecuteContainer(f.getDockerConfig(projectSubPath))
	if err != nil {
		return output, err
	}

	return output, f.parseOutput(output)
}

func (f *Formatter) getDockerConfig(projectSubPath string) *docker.AnalysisData {
	analysisData := &docker.AnalysisData{
		CMD:      f.AddWorkDirInCmd(CMD, projectSubPath, tools.Semgrep),
		Language: languages.Generic,
	}

	return analysisData.SetImage(f.GetCustomImageByLanguage(languages.Generic), images.Generic)
}

func (f *Formatter) parseOutput(output string) error {
	var analysis *sgAnalysis

	if err := json.Unmarshal([]byte(output), &analysis); err != nil {
		return err
	}

	for _, result := range analysis.Results {
		item := result
		f.AddNewVulnerabilityIntoAnalysis(f.newVulnerabilityFromResult(&item))
	}

	return nil
}

func (f *Formatter) newVulnerabilityFromResult(result *sgResult) *vulnerability.Vulnerability {
	vuln := &vulnerability.Vulnerability{
		RuleID:       result.CheckID,
		SecurityTool: tools.Semgrep,
		Details:      result.Extra.Message,
		Severity:     f.getSeverity(result.Extra.Severity),
		Line:         strconv.Itoa(result.Start.Line),
		Column:       strconv.Itoa(result.Start.Col),
		File:         result.Path,
		Code:         f.GetCodeWithMaxCharacters(result.Extra.Code, 0),
		Language:     f.getLanguageByFile(result.Path),
	}
	return f.SetCommitAuthor(vulnhash.Bind(vuln))
}

func (f *Formatter) getLanguageByFile(file string) languages.Language {
	if language, ok := f.getLanguagesMap()[filepath.Ext(file)]; ok {
		return language
	}
	return languages.Unknown
}

func (f *Formatter) getLanguagesMap() map[string]languages.Language {
	return map[string]languages.Language{
		".go":   languages.Go,
		".java": languages.Java,
		".js":   languages.Javascript,
		".jsx":  languages.Javascript,
		".tsx":  languages.Typescript,
		".ts":   languages.Typescript,
		".py":   languages.Python,
		".rb":   languages.Ruby,
		".c":    languages.C,
		".html": languages.HTML,
	}
}

func (f *Formatter) getSeverity(resultSeverity string) severities.Severity {
	switch resultSeverity {
	case "ERROR":
		return severities.High
	case "WARNING":
		return severities.Medium
	}
	return severities.Low
}
