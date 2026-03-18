/*
Copyright 2026 Scott Friedman

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/scttfrdmn/bagboy/pkg/packager"
	"github.com/scttfrdmn/bagboy/pkg/packager/appimage"
	"github.com/scttfrdmn/bagboy/pkg/packager/apptainer"
	"github.com/scttfrdmn/bagboy/pkg/packager/brew"
	"github.com/scttfrdmn/bagboy/pkg/packager/cargo"
	"github.com/scttfrdmn/bagboy/pkg/packager/chocolatey"
	"github.com/scttfrdmn/bagboy/pkg/packager/deb"
	"github.com/scttfrdmn/bagboy/pkg/packager/dmg"
	"github.com/scttfrdmn/bagboy/pkg/packager/docker"
	"github.com/scttfrdmn/bagboy/pkg/packager/flatpak"
	"github.com/scttfrdmn/bagboy/pkg/packager/installer"
	"github.com/scttfrdmn/bagboy/pkg/packager/msi"
	"github.com/scttfrdmn/bagboy/pkg/packager/msix"
	"github.com/scttfrdmn/bagboy/pkg/packager/nix"
	"github.com/scttfrdmn/bagboy/pkg/packager/npm"
	"github.com/scttfrdmn/bagboy/pkg/packager/pypi"
	"github.com/scttfrdmn/bagboy/pkg/packager/rpm"
	"github.com/scttfrdmn/bagboy/pkg/packager/scoop"
	"github.com/scttfrdmn/bagboy/pkg/packager/snap"
	"github.com/scttfrdmn/bagboy/pkg/packager/spack"
	"github.com/scttfrdmn/bagboy/pkg/packager/winget"
)

// newPackagerRegistry creates and populates a Registry with all 20 packagers.
func newPackagerRegistry() *packager.Registry {
	reg := packager.NewRegistry()
	reg.Register(brew.New())
	reg.Register(scoop.New())
	reg.Register(deb.New())
	reg.Register(rpm.New())
	reg.Register(chocolatey.New())
	reg.Register(winget.New())
	reg.Register(snap.New())
	reg.Register(appimage.New())
	reg.Register(flatpak.New())
	reg.Register(npm.New())
	reg.Register(pypi.New())
	reg.Register(docker.New())
	reg.Register(apptainer.New())
	reg.Register(dmg.New())
	reg.Register(msi.New())
	reg.Register(msix.New())
	reg.Register(cargo.New())
	reg.Register(nix.New())
	reg.Register(spack.New())
	reg.Register(installer.New())
	return reg
}
