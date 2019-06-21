# DNSPUMP
This tool helps you deliver files to machines entirely over DNS using `TXT` records. Point `ns` records at a box, host DNSPUMP and you're ready to roll.

DNSPUMP was hastily hacked together over the course of two days and relies entirely on the excellence of Miek Gieben's Go DNS package [miekg/dns](https://github.com/miekg/dns) [^dns].

[^dns]: https://miek.nl/2014/august/16/go-dns-package/

Check out the corresponding [post.](https://dualuse.io/blog/dnspump/)

# REQUIREMENTS

go get -u github.com/miekg/dns

# BUILDING

go build

# COMPLAINTS

Yes.

# LICENSE

MIT.