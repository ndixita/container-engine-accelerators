// Copyright 2017 Google Inc. All Rights Reserved.
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

package util

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/fsnotify/fsnotify"
)

func DeviceNameFromPath(path string) (string, error) {
	gpuPathRegex := regexp.MustCompile("/dev/(nvidia[0-9]+)$")
	m := gpuPathRegex.FindStringSubmatch(path)
	if len(m) != 2 {
		return "", fmt.Errorf("path (%s) is not a valid GPU device path", path)
	}
	return m[1], nil
}

// Files creates a Watcher for the specified files.
func Files(files ...string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		err = watcher.Add(f)
		if err != nil {
			watcher.Close()
			return nil, err
		}
	}
	return watcher, nil
}

func NUMANode(pciInfo nvml.PciInfo, pciDevicesRoot string) (numaEnabled bool, numaNode int, err error) {
	var bytesT []byte
	for _, b := range pciInfo.BusId {
		if byte(b) == '\x00' {
			break
		}
		bytesT = append(bytesT, byte(b))
	}

	// Discard leading zeros.
	busID := strings.ToLower(strings.TrimPrefix(string(bytesT), "0000"))

	numaNodeFile := fmt.Sprintf("%s/%s/numa_node", pciDevicesRoot, busID)
	// glog.Infof("Reading NUMA node information from %q", numaNodeFile)
	b, err := os.ReadFile(numaNodeFile)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read NUMA information from %v busID %q file: %v", pciInfo.BusId, numaNodeFile, err)
	}

	numaNode, err = strconv.Atoi(string(bytes.TrimSpace(b)))
	if err != nil {
		return false, 0, fmt.Errorf("eror parsing value for NUMA node: %v", err)
	}

	if numaNode < 0 {
		return false, 0, nil
	}

	return true, numaNode, nil
}
