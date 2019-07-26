package netconf

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

// BareMetalType defines the type of configuration to apply.
type BareMetalType int

const (
	// FileModeSystemd represents a file mode that allows systemd to read e.g. /etc/systemd/network files.
	FileModeSystemd = 0644
	// FileModeSixFourFour represents file mode 0644
	FileModeSixFourFour = 0644
	// FileModeDefault represents the default file mode sufficient e.g. to /etc/network/interfaces or /etc/frr.conf.
	FileModeDefault = 0600
	// Firewall defines the bare metal server to function as firewall.
	Firewall BareMetalType = iota
	// Machine defines the bare metal server to function as machine.
	Machine
)

type (
	// Configurator is an interface to configure bare metal servers.
	Configurator interface {
		Configure()
	}

	// CommonConfigurator contains information that is common to all configurators.
	CommonConfigurator struct {
		Kb KnowledgeBase
	}

	// MachineConfigurator is a configurator that configures a bare metal server as 'machine'.
	MachineConfigurator struct {
		CommonConfigurator
	}

	// FirewallConfigurator is a configurator that configures a bare metal server as 'firewall'.
	FirewallConfigurator struct {
		CommonConfigurator
	}
)

// NewConfigurator creates a new configurator.
func NewConfigurator(kind BareMetalType, kb KnowledgeBase) Configurator {
	var result Configurator
	switch kind {
	case Firewall:
		fw := FirewallConfigurator{}
		fw.CommonConfigurator = CommonConfigurator{kb}
		result = fw
	case Machine:
		m := MachineConfigurator{}
		m.CommonConfigurator = CommonConfigurator{kb}
		result = m
	default:
		log.Fatalf("Unknown kind of configurator: %v", kind)
	}
	return result
}

// Configure applies configuration to a bare metal server to function as 'machine'.
func (configurator MachineConfigurator) Configure() {
	applyCommonConfiguration(Machine, configurator.Kb)
}

// Configure applies configuration to a bare metal server to function as 'firewall'.
func (configurator FirewallConfigurator) Configure() {
	applyCommonConfiguration(Firewall, configurator.Kb)

	src := mustTmpFile("rules.v4_")
	validatorIPv4 := IptablesV4Validator{IptablesValidator{src}}
	applier := NewIptablesConfigApplier(configurator.Kb, validatorIPv4)
	applyAndCleanUp(applier, TplIptablesV4, src, "/etc/iptables/rules.v4", FileModeDefault)

	src = mustTmpFile("rules.v6_")
	validatorIPv6 := IptablesV6Validator{IptablesValidator{src}}
	applier = NewIptablesConfigApplier(configurator.Kb, validatorIPv6)
	applyAndCleanUp(applier, TplIptablesV6, src, "/etc/iptables/rules.v6", FileModeDefault)

	chrony, err := NewChronyServiceEnabler(configurator.Kb)
	if err != nil {
		log.Warnf("failed to configure Chrony: %v", err)
	} else {
		err := chrony.Enable()
		if err != nil {
			log.Errorf("enabling Chrony failed: %v", err)
		}
	}
}

func applyCommonConfiguration(kind BareMetalType, kb KnowledgeBase) {
	src := mustTmpFile("interfaces_")
	applier := NewIfacesConfigApplier(kind, kb, src)
	tpl := TplFirewallIfaces
	if kind == Machine {
		tpl = TplMachineIfaces
	}
	applyAndCleanUp(applier, tpl, src, "/etc/network/interfaces", FileModeDefault)

	src = mustTmpFile("hosts_")
	applier = NewHostsApplier(kb, src)
	applyAndCleanUp(applier, TplHosts, src, "/etc/hosts", FileModeDefault)

	src = mustTmpFile("hostname_")
	applier = NewHostnameApplier(kb, src)
	applyAndCleanUp(applier, TplHostname, src, "/etc/hostname", FileModeSixFourFour)

	src = mustTmpFile("frr_")
	applier = NewFrrConfigApplier(kind, kb, src)
	tpl = TplFirewallFRR
	if kind == Machine {
		tpl = TplMachineFRR
	}
	applyAndCleanUp(applier, tpl, src, "/etc/frr/frr.conf", FileModeDefault)

	for i, nic := range kb.Nics {
		prefix := fmt.Sprintf("lan%d_link_", i)
		src = mustTmpFile(prefix)
		applier = NewSystemdLinkApplier(kind, kb.Machineuuid, i, nic, src)
		dest := fmt.Sprintf("/etc/systemd/network/%d0-lan%d.link", i+1, i)
		applyAndCleanUp(applier, TplSystemdLink, src, dest, FileModeSystemd)

		prefix = fmt.Sprintf("lan%d_network_", i)
		src = mustTmpFile(prefix)
		applier = NewSystemdNetworkApplier(kb.Machineuuid, i, src)
		dest = fmt.Sprintf("/etc/systemd/network/%d0-lan%d.network", i+1, i)
		applyAndCleanUp(applier, TplSystemdNetwork, src, dest, FileModeSystemd)
	}
}

func applyAndCleanUp(applier network.Applier, tpl, src, dest string, mode os.FileMode) {
	log.Infof("rendering %s to %s (mode: %ui)", tpl, dest, mode)
	file := mustRead(tpl)
	mustApply(applier, file, src, dest)
	err := os.Chmod(dest, mode)
	if err != nil {
		log.Errorf("error to chmod %s to %ui", dest, mode)
	}
	_ = os.Remove(src)
}

func mustApply(applier network.Applier, tpl, src, dest string) {
	t := template.Must(template.New(TplFirewallIfaces).Parse(tpl))
	err := applier.Apply(*t, src, dest, false)
	if err != nil {
		log.Panic(err)
	}
}

func mustRead(name string) string {
	c, err := ioutil.ReadFile(name)
	if err != nil {
		log.Panic(err)
	}
	return string(c)
}

func mustTmpFile(prefix string) string {
	f, err := ioutil.TempFile("/etc/metal/networker/", prefix)
	if err != nil {
		log.Panic(err)
	}
	err = f.Close()
	if err != nil {
		log.Panic(err)
	}
	return f.Name()
}
