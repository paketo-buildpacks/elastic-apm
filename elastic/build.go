/*
 * Copyright 2018-2022 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package elastic

import (
	"fmt"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)
	result := libcnb.NewBuildResult()

	pr := libpak.PlanEntryResolver{Plan: context.Plan}

	dr, err := libpak.NewDependencyResolver(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency resolver\n%w", err)
	}

	dc, err := libpak.NewDependencyCache(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency cache\n%w", err)
	}
	dc.Logger = b.Logger

	if _, ok, err := pr.Resolve("elastic-apm-dotnet"); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve elastic-apm-dotnet plan entry\n%w", err)
	} else if ok {
		dep, err := dr.Resolve("elastic-apm-dotnet", "")
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		dna, be := NewDotNetAgent(dep, dc)
		dna.Logger = b.Logger
		result.Layers = append(result.Layers, dna)
		result.BOM.Entries = append(result.BOM.Entries, be)
	}

	if _, ok, err := pr.Resolve("elastic-apm-java"); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve elastic-apm-java plan entry\n%w", err)
	} else if ok {
		dep, err := dr.Resolve("elastic-apm-java", "")
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		ja, be := NewJavaAgent(dep, dc)
		ja.Logger = b.Logger
		result.Layers = append(result.Layers, ja)
		result.BOM.Entries = append(result.BOM.Entries, be)
	}

	if _, ok, err := pr.Resolve("elastic-apm-nodejs"); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve elastic-apm-nodejs plan entry\n%w", err)
	} else if ok {
		dep, err := dr.Resolve("elastic-apm-nodejs", "")
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		na, be := NewNodeJSAgent(context.Application.Path, dep, dc)
		na.Logger = b.Logger
		result.Layers = append(result.Layers, na)
		result.BOM.Entries = append(result.BOM.Entries, be)
	}

	h, be := libpak.NewHelperLayer(context.Buildpack, "properties")
	h.Logger = b.Logger
	result.Layers = append(result.Layers, h)
	result.BOM.Entries = append(result.BOM.Entries, be)

	return result, nil
}
