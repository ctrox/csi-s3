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

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ctrox/csi-s3/pkg/driver"
)

func init() {
	flag.Set("logtostderr", "true")
}

var (
	endpoint = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID   = flag.String("nodeid", "", "node id")
)

func main() {
	flag.Parse()

	// We're running in the container as PID-1 which gets some special
	// treatment by the kernel. In particular, if a process in the container
	// terminates and there are still active child processes, the kernel will move
	// those orphaned processes to be child processes of PID-1 and signal it
	// by sending a SIGCHLD. Init-systems are expected to handle this case by
	// reaping those "orphan" processes once they exit.
	//
	// Since all available mounters are instructed to daemonize, we need to reap
	// the daemonized processes since their parent (the mounter) exists once the daemon
	// is running.
	go func() {
		ch := make(chan os.Signal, 1)

		signal.Notify(ch, syscall.SIGCHLD)

		for range ch {
			var status syscall.WaitStatus
			pid, err := syscall.Wait4(-1, &status, 0, nil)
			if err != nil {
				// we might receive ECHILD when the mounter exits after daemonizing.
				// We'll be late calling Wait4 here as that process is already reaped
				// since we're using exec.Command().Run() which already calls Waitpid
				if val, ok := err.(syscall.Errno); !ok || val != syscall.ECHILD {
					log.Printf("failed to call wait4: %s\n", err)
				}

			} else {
				log.Printf("repeated child %d: status=%d\n", pid, status.ExitStatus())
			}
		}
	}()

	driver, err := driver.New(*nodeID, *endpoint)
	if err != nil {
		log.Fatal(err)
	}
	driver.Run()
	os.Exit(0)
}
