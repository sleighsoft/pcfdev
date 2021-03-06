package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"pcfdev/cert"
	"pcfdev/fs"
	"pcfdev/provisioner"
	"pcfdev/provisioner/commands"
	"strconv"
	"syscall"
	"time"
)

var (
	provisionScriptPath = "/var/pcfdev/run"
	timeoutInSeconds    = "3600"
	distro              = "pcf"
)

func main() {
	provisionTimeout, err := strconv.Atoi(timeoutInSeconds)
	if err != nil {
		fmt.Printf("Error: %s.", err)
		os.Exit(1)
	}

	p := &provisioner.Provisioner{
		Cert: &cert.Cert{},
		CmdRunner: &provisioner.ConcreteCmdRunner{
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
			Timeout: time.Duration(provisionTimeout) * time.Second,
		},
		FS: &fs.FS{},
		DisableUAAHSTS: &commands.DisableUAAHSTS{
			WebXMLPath: "/var/vcap/packages/uaa/tomcat/conf/web.xml",
		},
		ConfigureDnsmasq: &commands.ConfigureDnsmasq{
			Domain:     os.Args[1],
			ExternalIP: os.Args[2],
			FS:         &fs.FS{},
			CmdRunner: &provisioner.ConcreteCmdRunner{
				Stdout:  ioutil.Discard,
				Stderr:  ioutil.Discard,
				Timeout: time.Duration(provisionTimeout) * time.Second,
			},
		},

		Distro: distro,
	}

	if err := p.Provision(provisionScriptPath, os.Args[1:]...); err != nil {
		switch err.(type) {
		case *exec.ExitError:
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					os.Exit(status.ExitStatus())
				} else {
					os.Exit(1)
				}
			}
		case *provisioner.TimeoutError:
			fmt.Printf("Timed out after %s seconds.\n", timeoutInSeconds)
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
}
