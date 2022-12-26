# reverse-dns-enum
Do FQDN lookups across internal subnets to find hosts.

## Executables

Download `executables.zip` and extract.  Available executables:
reverse-dns-enum.exe
reverse-dns-enum.sh

## Run
Command line options
```azure
-d   (Required) Specify a local DNS server IP address. Example: 192.168.1.155
-n   (Required) CIDR notation of a newtork to scan. Example: 192.168.255.255/24
-t   (Optional) Number of threads.  Default: 1
```
### Linux

`./reverse-dns-enum -d 192.168.1.1 -n 192.168.1.255/24`

### Windows

`reverse-dns-enum -d 192.168.1.1 -n 192.168.1.255/24`

###

Example output:
```azure
Host found: [domain1.localdomain.] ip: 192.168.1.42
Host found: [domain2.localdomain.] ip: 192.168.1.76
Host found: [domain3.localdomain.] ip: 192.168.1.81
Time: 4.086481587s
Hostnames: [{"192.168.1.42":["domain1.localdomain."]},{"192.168.1.76":["domain2.localdomain."]},{"192.168.1.81":["domain3.localdomain."]}}]
```

### Note about threads

Depending on the number of IPs to check, this list will be divided by the number of threads specified and each "thread" will spin off a go routine to check off that chunk of IPs.
