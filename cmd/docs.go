package cmd

const (
	GInterface      = `The interface to operate on`
	GMatchBandwidth = `
Match abusers based on bandwidth consumption. The param allows
for a "floor:ceil"" value where you can match abusers who sit
within a bandwidth range (useful when adding multiple jails of
the same type) or you can add a "floor:" value where the ceil
part is ommited which will target every abuser that are over
the floor value (useful as a catch-all rule)

Expected params are integers where the value will be suffixed
with mbit: 100:200 = 100mbit:200mbit
`
	GMatchConnections = `Match abusers by the number of
connections. The param allows for a "floor:ceil"" value where
you can match abusers who sit within a bandwidth range (useful
when adding multiple jails of the same type) or you can add a
"floor:" value where the ceil part is ommited which will target
every abuser that are over the floor value (useful as a
catch-all rule)

Expected params are integers where the value is the number of
open connections you're targetting
`
	GPenaltyDrop      = `Matched connections will be dropped`
	GPenaltyBandwidth = `
Matched connections will be placed into a temporary restriction
table that limits their bandwidth to the specified amounts. The
param requires a "rate:ceil" value combo where "rate" is
mandatory and "ceil" optional.

Expected params are integers where the value will be suffixed
with mbit: 100:200 = 10mbit:1mbit
`
	GIdentifier = `The jail identifier.
To see a list of existing jails (and thier identifiers)
execute "jail list"`
	
	GCmdApplyShort     = "Applies configuration script"
	GCmdJailShort      = "Manages jails"
	GCmdJailDelShort   = "Deletes jail"
	GCmdJailAddShort   = "Creates a jail for criminals"
	GCmdJailsListShort = "Lists all jails in database"
	GCmdInspectShort   = "Will print the current script"
	
	GCmdApplyLong = `
This command will apply all iptables and traffic control rules.
THIS IS THE REAL DEAL so make sure you backed-up your iptables
rules (checkout "iptables-save") before running this command.

if you want to inspect the executed script before actually
aplying anything checkout the "inspect" command.

This project is in a very early stage so we might break some
things on certain systems. Play safe and smart - you have been
warned!
`
	
	GCmdInspectLong = `
This command prints the full list of os commands that
was/will be executed for this config. It just prints, without
executing anything. Useful when you want to debug things before
applying the commands.

When you're happy with the results and wish to apply use the
"apply" command
`
	
	GCmdJailAddLong = `
The jail is temporary place where abusive connections are being
placed. It is defined by a match and a penalty.

The match is just a set of parameters that helps us identify
abusers (like the number of open connections or connections
towards large downloads)

The penalty is the action taken once a match is identified. It
can be anything from dropping the connection or limiting the
bandwidth.
`
)
