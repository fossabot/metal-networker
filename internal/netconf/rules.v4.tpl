{{- /*gotype: git.f-i-ts.de/cloud-native/metal/metal-networker/internal/netconf.IptablesData*/ -}}
{{ .Comment }}

########################################################################################################################
# Default table definitions to handle:
# - packets destined to local sockets
# - packets routed through the box
# - locally-generated packets
#
*filter
# Allow any traffic by default.
:INPUT DROP [0:0]
:FORWARD ACCEPT [0:0]
:OUTPUT DROP [0:0]
:refuse - [0:0]

# Control behavior for incoming packets.
## Accept
--append INPUT --match conntrack --ctstate RELATED,ESTABLISHED --match comment --comment "stateful input" --jump ACCEPT
--append INPUT --in-interface lo --match comment --comment "BGP unnumbered" --jump ACCEPT
--append INPUT --in-interface lan0 --source 10.0.0.0/8 --protocol udp --match udp --destination-port 4789 --match comment --comment "incoming VXLAN lan0" --jump ACCEPT
--append INPUT --in-interface lan1 --source 10.0.0.0/8 --protocol udp --match udp --destination-port 4789 --match comment --comment "incoming VXLAN lan1" --jump ACCEPT
--append INPUT --protocol tcp --match tcp --destination-port 22 --match conntrack --ctstate NEW --jump ACCEPT --match comment --comment "SSH incoming connections"
## Drop
--append INPUT --match conntrack --ctstate INVALID --match comment --comment "drop invalid packets to prevent malicious activity" --jump DROP
--append INPUT --jump refuse

# Control behavior for forwarded packets.
## Drop
--append FORWARD --match conntrack --ctstate INVALID --match comment --comment "drop invalid packets from forwarding to prevent malicious activity" --jump DROP
--append FORWARD --protocol tcp --match tcp --destination-port 179 --match conntrack --ctstate NEW --match comment --comment "block bgp forward to machines" --jump refuse
## Accept

# Control behavior for outgoing packets.
--append OUTPUT --out-interface lo --match comment --comment "lo output required e.g. for chrony" --jump ACCEPT
--append OUTPUT --match conntrack --ctstate RELATED,ESTABLISHED --match comment --comment "stateful output"  --jump ACCEPT
--append OUTPUT --destination 10.0.0.0/8 --protocol udp --match udp --destination-port 4789 --match comment --comment "outgoing VXLAN" --jump ACCEPT
# Drop
--append OUTPUT --match conntrack --ctstate INVALID --match comment --comment "drop invalid packets" --jump DROP
--append OUTPUT --jump refuse

# Control behavior to handle unwanted traffic.
# The refuse chain logs the package with a delay to avoid flooding.
--append refuse --match limit --limit 2/min --jump LOG --log-prefix "iptables-dropped: "
# Drop the package after having it logged to refuse it.
--append refuse --jump DROP

COMMIT
# END OF *filter #######################################################################################################

########################################################################################################################
# Consulted when a packet that creates a new connection is encountered.
*nat
:PREROUTING ACCEPT [0:0]
:INPUT ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
:POSTROUTING ACCEPT [0:0]
{{- range .SNAT }}
    {{- $cmt:=.Comment }}
    {{- $out:=.OutInterface }}
    {{- range .SourceSpecs }}
--append POSTROUTING --source {{ . }} --out-interface {{ $out }} --match comment --comment "{{ $cmt }}" --jump MASQUERADE
    {{- end }}
{{- end }}

COMMIT
# END OF *nat ##########################################################################################################
