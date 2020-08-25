## udns

A barebones, simple DNS server implementation for edification and use with
my Wireguard VPN.

I've written a [blog post](https://blog.aos.sh/2020/08/23/under-the-hood-of-a-simple-dns-server/)
about the implementation of this DNS server.

## Requirements

- `go` (tested on `1.14`, but could/should work on older versions)

## Installation & Usage

1. Clone repository
2. Build with `go build`
3. Run with:
```
./udns [-port 8053] [-zonefile file.zone] [-address "127.0.0.1"] [-forward-server "1.1.1.1:53"]
```

- Port defaults to `8053`
- Zonefile defaults to `master.zone` (you can use http://zonefile.org to create a zonefile)
- Address defaults to empty
- Forward server defaults to `1.1.1.1:53`

There is also a bundled `systemd` unit file that can be modified and copied
into `/etc/systemd/system/` for an auto-starting service.

### A note about Linux distributions with `systemd`

If you are using `systemd-resolved`, keep in mind that you will not be able to
run this with an empty address on port `53`. This is because
`systemd-resolved` creates a local DNS server/cache/resolver that
listens on `127.0.0.53:53` and an empty address attempts to listen on all
availabe IP addresses of the local system. To get around this, you can specify
an `-address` for others to listen on, a simple one being `127.0.0.1`.
