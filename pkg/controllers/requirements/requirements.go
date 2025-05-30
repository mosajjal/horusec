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

package requirements

import (
	"github.com/mosajjal/horusec/pkg/controllers/requirements/docker"
	"github.com/mosajjal/horusec/pkg/controllers/requirements/git"
)

type ValidationFn func() error

type Requirements struct {
	gitValidationFn    ValidationFn
	dockerValidationFn ValidationFn
}

func NewRequirements() *Requirements {
	return &Requirements{
		gitValidationFn:    git.Validate,
		dockerValidationFn: docker.Validate,
	}
}

func (r *Requirements) ValidateDocker() error {
	return r.dockerValidationFn()
}

func (r *Requirements) ValidateGit() error {
	return r.gitValidationFn()
}
