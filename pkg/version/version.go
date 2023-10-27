// Copyright 2023 The Simila Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package version

import "fmt"

// Version is the acr app version. To be injected.
var Version string

// GitCommit is git commit used for the build
var GitCommit string

// BuildDate is the date when current binary was built
var BuildDate string

// GoVersion is the used Go version
var GoVersion string

// BuildVersionString returns the information about the application version
func BuildVersionString() string {
	if Version == "" {
		return "<version is not set>"
	}
	return fmt.Sprintf("ariadne  %s %s %s. Built with %s", Version, GitCommit, BuildDate, GoVersion)
}
