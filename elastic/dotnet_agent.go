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
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type DotNetAgent struct {
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewDotNetAgent(dependency libpak.BuildpackDependency, cache libpak.DependencyCache) (DotNetAgent, libcnb.BOMEntry) {
	contributor, entry := libpak.NewDependencyLayer(dependency, cache, libcnb.LayerTypes{
		Launch: true,
	})
	return DotNetAgent{LayerContributor: contributor}, entry
}

func (dn DotNetAgent) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	dn.LayerContributor.Logger = dn.Logger

	return dn.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		dn.Logger.Bodyf("Copying to %s", layer.Path)

		bn := filepath.Base(artifact.Name())
		dir := strings.TrimSuffix(bn, filepath.Ext(bn))
		zd, err := os.MkdirTemp("", dir)
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to create temp dir %s for unzipping %s\n%w", zd, artifact.Name(), err)
		}
		dest := filepath.Join(layer.Path, dir)

		if err := os.Chmod(zd, 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to change permissions on %s\n%w", zd, err)
		}

		if err := Unzip(artifact.Name(), zd); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to unzip %s to %s\n%w", artifact.Name(), zd, err)
		}

		if err := filepath.WalkDir(zd, changePermissions); err != nil {
			return libcnb.Layer{}, fmt.Errorf("Could not chmod recursively %s\n%w", zd, err)
		}

		if err := sherpa.CopyDir(zd, dest); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy %s to %s\n%w", zd, dest, err)
		}

		layer.LaunchEnvironment.Appendf("DOTNET_STARTUP_HOOKS", ":", "%s/ElasticApmAgentStartupHook.dll", dest)

		return layer, nil
	})
}

func (dn DotNetAgent) Name() string {
	return dn.LayerContributor.LayerName()
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func changePermissions(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	return os.Chmod(path, 0755)
}
