// Copyright 2021 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
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

package dotnetcli

import (
	"path/filepath"
	"strings"

	"github.com/ZupIT/horusec-devkit/pkg/entities/vulnerability"
	"github.com/ZupIT/horusec-devkit/pkg/enums/confidence"
	"github.com/ZupIT/horusec-devkit/pkg/enums/languages"
	"github.com/ZupIT/horusec-devkit/pkg/enums/tools"
	"github.com/ZupIT/horusec-devkit/pkg/utils/logger"

	"github.com/mosajjal/horusec/pkg/entities/docker"
	"github.com/mosajjal/horusec/pkg/enums/images"
	"github.com/mosajjal/horusec/pkg/helpers/messages"
	"github.com/mosajjal/horusec/pkg/services/formatters"
	"github.com/mosajjal/horusec/pkg/utils/file"
	vulnhash "github.com/mosajjal/horusec/pkg/utils/vuln_hash"
)

const CsProjExt = ".csproj"

type Formatter struct {
	formatters.IService
}

func NewFormatter(service formatters.IService) *Formatter {
	return &Formatter{
		IService: service,
	}
}

func (f *Formatter) StartAnalysis(projectSubPath string) {
	if f.ToolIsToIgnore(tools.DotnetCli) || f.IsDockerDisabled() {
		logger.LogDebugWithLevel(messages.MsgDebugToolIgnored + tools.DotnetCli.ToString())
		return
	}

	output, err := f.startDotnetCli(projectSubPath)
	f.SetAnalysisError(err, tools.DotnetCli, output, projectSubPath)
	f.LogDebugWithReplace(messages.MsgDebugToolFinishAnalysis, tools.DotnetCli, languages.CSharp)
}

func (f *Formatter) startDotnetCli(projectSubPath string) (string, error) {
	f.LogDebugWithReplace(messages.MsgDebugToolStartAnalysis, tools.DotnetCli, languages.CSharp)

	output, err := f.checkOutputErrors(f.ExecuteContainer(f.getConfigData(projectSubPath)))
	if err != nil {
		return output, err
	}

	f.parseOutput(output, projectSubPath)
	return output, nil
}

func (f *Formatter) getConfigData(projectSubPath string) *docker.AnalysisData {
	analysisData := &docker.AnalysisData{
		CMD: f.AddWorkDirInCmd(
			CMD,
			file.GetSubPathByExtension(f.GetConfigProjectPath(), projectSubPath, "*.sln"), tools.DotnetCli,
		),
		Language: languages.CSharp,
	}

	return analysisData.SetImage(f.GetCustomImageByLanguage(languages.CSharp), images.Csharp)
}

func (f *Formatter) parseOutput(output, projectSubPath string) {
	if f.isInvalidOutput(output) {
		return
	}

	startIndex := strings.Index(output, separator)
	if startIndex < 0 {
		startIndex = 0
	}

	f.addVulnIntoAnalysis(output, projectSubPath, startIndex)
}

func (f *Formatter) addVulnIntoAnalysis(output, projectSubPath string, startIndex int) {
	for _, value := range strings.Split(output[startIndex:], separator) {
		dependency := f.parseDependencyValue(value)
		if dependency != nil && *dependency != (dotnetDependency{}) {
			vuln, err := f.newVulnerability(dependency, projectSubPath)
			if err != nil {
				f.SetAnalysisError(err, tools.DotnetCli, err.Error(), "")
				logger.LogErrorWithLevel(messages.MsgErrorGetDependencyCodeFilepathAndLine, err)
				continue
			}
			f.AddNewVulnerabilityIntoAnalysis(vuln)
		}
	}
}

func (f *Formatter) parseDependencyValue(value string) *dotnetDependency {
	dependency := new(dotnetDependency)

	for index, fieldValue := range f.formatOutput(value) {
		if strings.TrimSpace(fieldValue) == "" || strings.TrimSpace(fieldValue) == "\n" {
			continue
		}

		f.parseFieldByIndex(index, fieldValue, dependency)
	}

	return dependency
}

func (f *Formatter) formatOutput(value string) (result []string) {
	value = strings.ReplaceAll(value, "\n", "")
	value = strings.ReplaceAll(value, "\r", "")

	// TODO(matheus): We should not use color characters to split the value.
	// We should find a better approach here.
	for _, field := range strings.Split(value, "\u001B[39;49m") {
		field = strings.TrimSpace(field)
		if field != "" && strings.TrimSpace(field) != autoReferencedPacket {
			result = append(result, field)
		}
	}

	return result
}

func (f *Formatter) parseFieldByIndex(index int, fieldValue string, dependency *dotnetDependency) {
	switch index {
	case indexDependencyName:
		dependency.setName(fieldValue)
	case indexDependencyVersion:
		dependency.setVersion(fieldValue)
	case indexDependencySeverity:
		dependency.setSeverity(fieldValue)
	case indexDependencyDescription:
		dependency.setDescription(fieldValue)
	}
}

// nolint:funlen // This method will mount the vulnerability struct and is necessary more lines
func (f *Formatter) newVulnerability(dependency *dotnetDependency,
	projectSubPath string) (*vulnerability.Vulnerability, error,
) {
	dependencyInfo, err := file.GetDependencyCodeFilepathAndLine(f.GetConfigProjectPath(), projectSubPath,
		[]string{dependency.Name}, CsProjExt,
	)
	if err != nil {
		return nil, err
	}

	relFilePath, err := filepath.Rel(f.GetConfigProjectPath(), dependencyInfo.Path)
	if err != nil {
		return nil, err
	}

	vuln := &vulnerability.Vulnerability{
		SecurityTool: tools.DotnetCli,
		Language:     languages.CSharp,
		Confidence:   confidence.High,
		RuleID:       vulnhash.HashRuleID(dependency.getDescription()),
		Details:      dependency.getDescription(),
		Code:         dependencyInfo.Code,
		File:         f.removeHorusecFolder(relFilePath),
		Line:         dependencyInfo.Line,
		Severity:     dependency.getSeverity(),
	}

	vuln.DeprecatedHashes = f.getDeprecatedHashes(dependencyInfo.Path, *vuln)

	return f.SetCommitAuthor(vulnhash.Bind(vuln)), nil
}

func (f *Formatter) checkOutputErrors(output string, err error) (string, error) {
	if err != nil {
		return output, err
	}

	if strings.Contains(output, solutionNotFound) {
		return output, ErrorSolutionNotFound
	}

	return output, nil
}

func (f *Formatter) isInvalidOutput(output string) bool {
	return !(strings.Contains(output, "Top-level Package") && strings.Contains(output, "Requested") &&
		strings.Contains(output, "Resolved") && strings.Contains(output, "Severity"))
}

func (f *Formatter) removeHorusecFolder(path string) string {
	return filepath.Clean(strings.ReplaceAll(path, filepath.Join(".horusec", f.GetAnalysisID()), ""))
}

// getDeprecatedHashes necessary due a change that from now the hash is generated from a relative path, not an absolute
// path. This func exists to keep generating this old hashes with the absolute path and avoid some breaking changes.
// TODO: This will be removed after the release v2.10.0 be released
// nolint:gocritic // it has to be without pointer
func (f *Formatter) getDeprecatedHashes(absFilePath string, vuln vulnerability.Vulnerability) []string {
	vuln.File = f.removeHorusecFolder(absFilePath)

	return vulnhash.Bind(&vuln).DeprecatedHashes
}
