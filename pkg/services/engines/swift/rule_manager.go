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

package swift

import (
	engine "github.com/ZupIT/horusec-engine"

	"github.com/mosajjal/horusec/pkg/services/engines"
)

func NewRules() *engines.RuleManager {
	return engines.NewRuleManager(rules(), extensions())
}

func extensions() []string {
	return []string{".swift"}
}

func rules() []engine.Rule {
	return []engine.Rule{
		// And rules
		NewWeakCommonDesCryptoCipher(),
		NewWeakIDZDesCryptoCipher(),
		NewWeakBlowfishCryptoCipher(),
		NewWeakMD5CryptoCipher(),
		NewReverseEngineering(),
		NewTLS13NotUsed(),
		NewDTLS12NotUsed(),
		NewCoreDataDatabase(),
		// NewSQLiteDatabase(),

		// Or rules
		NewWeakDesCryptoCipher(),
		NewLoadHTMLString(),
		NewJailbreakDetect(),
		NewSha1Collision(),
		NewMD5Collision(),
		NewMD6Collision(),

		// Regular rules
		NewMD2Collision(),
		NewMD4Collision(),
		NewWebViewSafari(),
		NewFileProtection(),
		NewUIPasteboard(),
		NewKeyboardCache(),
		NewTLSMinimum(),
		NewRealmDatabase(),
		NewSQLInjection(),
	}
}
