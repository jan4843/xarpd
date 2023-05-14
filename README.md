# xarpd

xarpd bridges ARP requests between separate networks.

xarpd listens for incoming ARP requests, and when a request matches a configured subnetwork, it asks other peers running xarpd to resolve the request. When a peer can resolve the ARP request, xarpd responds to the initial ARP request with the hardware address of the host it is running on, so that traffic can be forwarded through it.

## Motivation

### Scenario

- `network1` is on `10.0.0.0/24` and has the hosts:
    - `server1` with IP address `10.0.0.1`
    - `device1` with IP address `10.0.0.42`
- `network2` is on `10.0.0.0/24` and has the hosts:
    - `server2` with IP address `10.0.0.129`
    - `device2` with IP address `10.0.0.170`

`network1` and `network2` have the same address but are different networks.

`server1` has a route to the subnetwork `10.0.0.128/25` of `network2` via a VPN and can therefore reach `server2` and `device2`.

It is assumed that `server1` and `server2` have IP forwarding enabled.

### Goal

`device1` needs to reach `device2` without the ability to join VPNs or configure custom routes. Therefore, the traffic would need to flow:
`device1` &rarr; `server1` &rarr; `server2` &rarr; `device2`.

### Solution

xarpd running on `server1` intercepts ARP requests from `device1` for `10.0.0.129` and replies with the hardware address of `server1`. Before replying, it asks `server2` to check whether `10.0.0.129` exists on `network2`.

## Usage

xarpd must be running on at least two networks to be useful.

On `server1` with IP address `10.0.0.1/24` in `network1`:

```console
$ xarpd 10.0.0.129/25
Forwarding ARP requests for 10.0.0.128/25 to 10.0.0.129
Listening for ARP on eth0
Listening for HTTP on 10.0.0.1:2707
```

On `server2` with IP address `10.0.0.129/24` in `network2`:

```console
$ xarpd 10.0.0.1/25
Forwarding ARP requests for 10.0.0.0/25 to 10.0.0.1
Listening for ARP on eth0
Listening for HTTP on 10.0.0.129:2707
```
