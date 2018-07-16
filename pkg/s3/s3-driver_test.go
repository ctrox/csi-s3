/*
Copyright 2017 The Kubernetes Authors.

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

package s3

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/kubernetes-csi/csi-test/pkg/sanity"
)

func TestDriver(t *testing.T) {
	socket := "/tmp/csi.sock"
	endpoint := "unix://" + socket

	if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove unix domain socket file %s, error: %s", socket, err)
	}
	cfg := &Config{
		AccessKeyID:     "FJDSJ",
		SecretAccessKey: "DSG643HGDS",
		Endpoint:        "http://127.0.0.1:9000",
	}
	driver, err := NewS3("test-node", endpoint, cfg)
	if err != nil {
		log.Fatal(err)
	}
	go driver.Run()

	mntDir, err := ioutil.TempDir("", "mnt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mntDir)

	sanityCfg := &sanity.Config{
		TargetPath: mntDir,
		Address:    endpoint,
	}

	sanity.Test(t, sanityCfg)
}
