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

package elastic_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libpak"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/elastic-apm/v4/elastic"
)

func testDotNetAgent(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Layers.Path, err = ioutil.TempDir("", "dotnet-agent-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("contributes .Net agent", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-elastic-apm-agent.zip",
			SHA256: "2dfe54fd4bb7589b2a3b63d80b4d8bd74f83cd8ae59dc6a70ad9da901d8c556d",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		dn, _ := elastic.NewDotNetAgent(dep, dc)
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = dn.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Launch).To(BeTrue())
		Expect(filepath.Join(layer.Path, "stub-elastic-apm-agent")).To(BeADirectory())
		Expect(filepath.Join(layer.Path, "stub-elastic-apm-agent", "ElasticApmAgentStartupHook.dll")).To(BeARegularFile())
		Expect(layer.LaunchEnvironment["DOTNET_STARTUP_HOOKS.delim"]).To(Equal(":"))
		Expect(layer.LaunchEnvironment["DOTNET_STARTUP_HOOKS.append"]).To(Equal(fmt.Sprintf("%s",
			filepath.Join(layer.Path, "stub-elastic-apm-agent\\ElasticApmAgentStartupHook.dll"))))
	})
}
