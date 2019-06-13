// Copyright 2019 Shift Cryptosecurity AG, Switzerland.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// BitBox Base Supervisor
// ----------------------
// monitors systemd or file logs to detect potential issues and take action.
//
// Functionality to implement:
// * System
//   - temperature control: monitor bbbfancontrol and throttle CPU if needed
//   - disk space: monitor free space on rootfs and ssd, perform cleanup of temp & logs
//   - swap: detect issues with swap file, no memory left or "zram decompression failed", perform reboot
//
// * Middleware
//   - monitor service availability
//
// * Bitcoin Core
//   - monitor service availability
//   - perform backup tasks
//   - switch between IBD and normal operation mode (e.g. adjust dbcache)
//
// * c-lightning
//   - monitor service availability
//   - perform backup tasks (once possible)
//
// * electrs
//   - monitor service availability
//   - track initial sync and full database compaction, restart service if needed
//
// * NGINX, Grafana, ...
//   - monitor service availability
//
//

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/coreos/go-systemd/sdjournal"
)

//Q: how to create this "Follower" according to best practice?
func startFollower(service string, logline chan string) {
	fmt.Println(service + ": started")

	logline <- "channel: " + service + ": started"

	jconf := sdjournal.JournalReaderConfig{
		Since: time.Duration(-15) * time.Second,
		Matches: []sdjournal.Match{
			{
				Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT,
				Value: service,
			},
		},
	}

	jr, err := sdjournal.NewJournalReader(jconf)

	if err != nil {
		panic(err)
	}

	if jr == nil {
		fmt.Println(service + ": got a nil reader")
		return
	}

	defer jr.Close()

	// Q: how to implement and use a Writer that pipes the journal entries to the `logline` channel?
	jr.Follow(nil, os.Stdout)
}

// Test only, make some beeps
func test(logline chan string) {
	for {
		time.Sleep(time.Duration(time.Second))
		logline <- "beep"
	}
}

func monitorJournal(logline chan string) {
	for {
		// endless loop
		fmt.Println(<-logline)
	}
}

func main() {
	versionNum := 0.1

	// parse command line arguments
	version := flag.Bool("version", false, "return program version")
	flag.Parse()

	fmt.Println("bbbsupervisor version", versionNum)
	if *version {
		os.Exit(0)
	}

	// monitoring routine and channel to process input from systemd followers
	logline := make(chan string)
	go monitorJournal(logline)

	// follower routines for systemd services
	go startFollower("NetworkManager.service", logline)
	go startFollower("user@1001.service", logline)
	go startFollower("kernel.service", logline)

	// make some beeps
	go test(logline)

	for {
		// endless loop
		// not sure yet what to put here
	}
}
