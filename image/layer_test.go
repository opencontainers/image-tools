// Copyright 2016 The Linux Foundation
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

package image

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateFilesystemChangeset(t *testing.T) {
	// create base fileystem
	tmp1, err := ioutil.TempDir("", "test-layer")
	if err != nil {
		t.Fatal(err)
	}
	//defer os.RemoveAll(tmp1)
	basepath := filepath.Join(tmp1, "base")
	err = os.MkdirAll(basepath, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// get the file list of base layer
	expectedFiles := map[string]bool{
		"bin/":                    true,
		"bin/app":                 true,
		"bin/test":                true,
		"bin/tool":                true,
		"etc/":                    true,
		"etc/app.cfg":             true,
		"etc/tool.cfg.d/":         true,
		"etc/tool.cfg.d/tool.cfg": true,
	}
	err = createFilesystem(basepath, expectedFiles, nil)
	if err != nil {
		t.Fatal(err)
	}
	// create base layer
	tarfile := filepath.Join(tmp1, "base.tar")
	err = createLayer(basepath, "", tarfile)
	if err != nil {
		t.Fatal(err)
	}

	files, err := listTarFiles(tarfile)
	if err != nil {
		t.Fatal(err)
	}
	// verify base layer has packed all the file
	// and no expected file are packed.
	err = verify(expectedFiles, files)
	if err != nil {
		t.Fatal(err)
	}

	// create a identical copy of base
	snapshot1path := filepath.Join(tmp1, "base.s1")
	cpCmd := exec.Command("cp", "-a", basepath, snapshot1path)
	err = cpCmd.Run()
	if err != nil {
		t.Fatalf("fail to cp: %v", err)
	}

	// delete some file and add some file in base.s1
	layer1FilesDiff := map[string]bool{
		"bin/tool":              false,
		"bin/app.tool":          true,
		"etc/app.cfg":           false,
		"etc/tool.cfg.d/":       false,
		"etc/app.cfg.d/":        true,
		"etc/app.tool.cfg":      true,
		"etc/app.cfg.d/app.cfg": true,
	}
	modifiedfiles := map[string]func(string) error{
		"bin/app": func(path string) error {
			return ioutil.WriteFile(path, []byte(fmt.Sprintf("Hello world")), 0755)
		},
	}

	err = createFilesystem(snapshot1path, layer1FilesDiff, modifiedfiles)
	if err != nil {
		t.Fatal(err)
	}

	// create layer diff
	tarfile1 := filepath.Join(tmp1, "base.s1.tar")
	err = createLayer(snapshot1path, basepath, tarfile1)
	if err != nil {
		t.Fatal(err)
	}

	execptedfiles := map[string]bool{
		"bin/":                  true,
		"bin/app":               true,
		"bin/app.tool":          true,
		"bin/.wh.tool":          true,
		"etc/":                  true,
		"etc/.wh.app.cfg":       true,
		"etc/app.cfg.d/":        true,
		"etc/app.cfg.d/app.cfg": true,
		"etc/app.tool.cfg":      true,
		"etc/.wh.tool.cfg.d":    true,
	}

	actualfile, err := listTarFiles(tarfile1)
	if err != nil {
		t.Fatal(err)
	}

	// verify if CreateLayer has created expected filesystem changeset
	err = verify(execptedfiles, actualfile)
	if err != nil {
		t.Fatal(err)
	}
}

func createLayer(child, parent, dest string) error {
	filename := filepath.Clean(dest)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	out, err := CreateLayer(child, parent)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, out)
	return err
}

func createFilesystem(path string, files map[string]bool, modify map[string]func(string) error) error {
	for f, add := range files {
		// add file
		if add {
			// create a directory
			if strings.HasSuffix(f, "/") {
				err := os.MkdirAll(filepath.Join(path, f), 0700)
				if err != nil {
					return err
				}
			} else { // create file
				file := filepath.Join(path, f)
				err := os.MkdirAll(filepath.Dir(file), 0700)
				if err != nil {
					return err
				}
				_, err = os.Create(file)
				if err != nil {
					return err
				}
			}
		} else { // remove file
			file := filepath.Join(path, f)
			err := os.RemoveAll(file)
			if err != nil {
				return err
			}
		}
	}

	// apply file modify
	for f, fun := range modify {
		err := fun(filepath.Join(path, f))
		if err != nil {
			return err
		}
	}
	return nil
}

func verify(m1 map[string]bool, m2 map[string]bool) error {
	for f := range m1 {
		if _, ok := m2[f]; !ok {
			return fmt.Errorf("expected file %v not exist", f)
		}
	}

	for f := range m2 {
		if _, ok := m1[f]; !ok {
			return fmt.Errorf("%v is not an expected file", f)
		}
	}
	return nil
}

func listTarFiles(path string) (map[string]bool, error) {
	var files = make(map[string]bool)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	tr := tar.NewReader(file)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		files[hdr.Name] = true
	}
	return files, nil
}
