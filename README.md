# metal-networker
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fmetal-stack%2Fmetal-networker.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fmetal-stack%2Fmetal-networker?ref=badge_shield)


Configures networking related resources such as interfaces, frr and nftables.

## Preconditions

Ubuntu operating system in place with the following packages installed: 

- ifupdown2
- frr > 7.0
- nftables


## Usage

metal-networker must be invoked with the configuration file as argument. It is expected that the configuration file 
contains valid YAML. See [./internal/netconf/testdata/firewall.yaml](internal/netconf/testdata/firewall.yaml) for a valid configuration for firewalls and [./internal/netconf/testdata/machine.yaml](internal/netconf/testdata/machine.yaml) for a valid configuration for machines.

```bash
# metal-networker <config-file>
./metal-networker machine|firewall configure --input install.yaml

```


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fmetal-stack%2Fmetal-networker.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fmetal-stack%2Fmetal-networker?ref=badge_large)