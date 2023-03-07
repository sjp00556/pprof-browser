// Copyright 2018 The Chubao Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"flag"
	"fmt"
	syslog "log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"github.com/sjp00556/pprof-browser/pkg/consts"
	"github.com/sjp00556/pprof-browser/pkg/server"
	"github.com/sjp00556/pprof-browser/pkg/util/config"
	sysutil "github.com/sjp00556/pprof-browser/pkg/util/sys"

	"github.com/jacobsa/daemonize"
)

var (
	configPort       = flag.String("p", "8888", "port")
	configDir        = flag.String("dir", "./", "work dir")
	configForeground = flag.Bool("f", false, "run foreground")
)

func interceptSignal(s server.Server) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	syslog.Println("action[interceptSignal] register system signal.")
	go func() {
		sig := <-sigC
		syslog.Printf("action[interceptSignal] received signal: %s.", sig.String())
		s.Shutdown()
	}()
}

func main() {
	flag.Parse()
	if *configDir == "" {
		fmt.Printf("use -dir to set files dir\n")
		os.Exit(1)
	}

	var err error
	if _, err = os.Stat(*configDir); err != nil {
		fmt.Printf("stat %v %v\n", *configDir, err)
		os.Exit(1)
	}

	if *configDir, err = filepath.Abs(*configDir); err != nil {
		fmt.Printf("get abs dir failed", err)
		os.Exit(1)
	}

	if !*configForeground {
		if err := startDaemon(); err != nil {
			fmt.Printf("Server start failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	/*
	 * We are in daemon from here.
	 * Must notify the parent process through SignalOutcome anyway.
	 */

	// Init server
	s := server.NewAPIServer()

	// Init output file
	stdoutLog := path.Join(*configDir, "stdout.log")
	outputFile, err := os.OpenFile(stdoutLog, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		err = fmt.Errorf("Fatal: failed to open output path %v - %v", stdoutLog, err)
		fmt.Println(err)
		daemonize.SignalOutcome(err)
		os.Exit(1)
	}
	defer func() {
		outputFile.Sync()
		outputFile.Close()
	}()
	syslog.SetOutput(outputFile)

	if err = sysutil.RedirectFD(int(outputFile.Fd()), int(os.Stderr.Fd())); err != nil {
		err = fmt.Errorf("Fatal: failed to redirect Stderr fd - %v", err)
		syslog.Println(err)
		daemonize.SignalOutcome(err)
		os.Exit(1)
	}

	if err = sysutil.RedirectFD(int(outputFile.Fd()), int(os.Stdout.Fd())); err != nil {
		err = fmt.Errorf("Fatal: failed to redirect Stdout fd - %v", err)
		syslog.Println(err)
		daemonize.SignalOutcome(err)
		os.Exit(1)
	}

	syslog.Printf("pprofweb started\n")

	interceptSignal(s)
	cfg := config.LoadConfig(map[string]interface{}{
		consts.CfgKeyPort: *configPort,
		consts.CfgKeyDir:  *configDir,
	})
	err = s.Start(cfg)
	if err != nil {
		err = fmt.Errorf("Fatal: failed to start the pprofweb daemon err %v - ", err)
		syslog.Println(err)
		daemonize.SignalOutcome(err)
		os.Exit(1)
	}

	daemonize.SignalOutcome(nil)

	// Block main goroutine until server shutdown.
	s.Sync()
	os.Exit(0)
}

func startDaemon() error {
	cmdPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("startDaemon failed: cannot get absolute command path, err(%v)", err)
	}

	args := []string{"-f"}
	args = append(args, "-p", *configPort)
	args = append(args, "-dir", *configDir)

	env := []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
	}

	err = daemonize.Run(cmdPath, args, env, os.Stdout)
	if err != nil {
		return fmt.Errorf("startDaemon failed: daemon start failed, cmd(%v) args(%v) env(%v) err(%v)\n", cmdPath, args, env, err)
	}

	return nil
}
