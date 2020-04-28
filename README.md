### WARNING: alpha software
I keep placing these warnings in code and everywhere else but playing with `tc` 
and `iptables` is no joke and this program has the potential to shut the
lights out on your machine when it comes to internet traffic of course.
 
## About
Simple way of shaping traffic on your server. You can add rules/jails that
come in the form of a `match` and a `penalty` to identify and punish abusers
respectively.

When identifying abusers the following match types are available:

- by number of open connections
- by consumed bandwidth

The possible applied penalties are:

- drop the connection
- limit bandwidth

### Install

`sh -c 'sh -c "$(curl -sL https://raw.githubusercontent.com/ciokan/shaper/master/install.sh)"'`

### Usage

__Scenario__: Identify users that have more than 100 connections open and
place them into a jail (bucket) where internet speed is capped at `1mbit`:

`./shaper jail add --match-connections=100: --penalty-bandwidth=1`

__Scenario__: Identify users that have more than 100 connections open and
drop any other connection above that limit:

`./shaper jail add --match-connections=100: --penalty-drop`

__Scenario__: Identify users performing downloads that have exceeded 10Mb in
 size and place them into a jail (bucket) where internet speed is capped at
  `1mbit`:
  
`./shaper jail add --match-size=10000000: --penalty-bandwidth=1`

__Scenario__: Identify users performing downloads that have exceeded 10Mb in
 size and drop their connections:
 
`./shaper jail add --match-size=10000000: --penalty-drop`

### Good to know

All values for bandwidth related matches or penalties are translated in `mbit
`. If you specify a `10` it means `10mbit`

The executable has `-h` help commands for every instruction so make sure to
check it out.
 
No command is executed unless you call `apply`. This is to allow you to inspect
(using `inspect`) what is about to be executed and make sure it looks ok.