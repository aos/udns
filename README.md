## udns

A barebones, simple DNS server implementation for edification and use with
my Wireguard VPN.

### Terminology

1. **Resource Record**: the basic building block of DNS. This is the data 
   requested by a client (eg. an address from a hostname) and answered by a DNS
   server.

2. **Zone**: A part of the domain space where a name server is considered to
   hold authoritative information. This information is usually listed in a
   "master" file known as a zone file.

3. **Client**: A client is any program that requests name server information,
   through a query that contains a query _name_ (QNAME), query _class_
   (QCLASS), and query _type_ (QTYPE).

3. Query types (**QTYPE**):
    - A (AAAA): IPv4 address, and IPv6 address (respectively)
    - CNAME: alias (canonical name)
    - PTR: reverse A (from address to hostname)
    - NS: name server
    - MX: mail exchange
    - SOA: Start of Authority, contains administrative information about the zone
    - HINFO: identifies CPU and OS used by host

4. Query classes (**QCLASS**):
    - IN: stands for INternet, the most common and widely used
    - CH: CHaos system

5. **Resolver**: programs that interface clients to DNS servers. That is,
   they receive a name request from a client, retrieve desired information
   from DNS servers, and return it in a format compatible with the program.
    - Stub resolvers: a simpler form of a resolver that moves the resolution
      function completely out of the local machine. It does no- or local-only caching.

### Resource Records

In a zone file, the contents of a resource record has the following format:

```
catcoffeecode.club.     1799    IN      A       157.245.253.239
        ^                 ^     ^       ^             ^
Query name (QNAME)       TTL    QCLASS  QTYPE       Answer
```
