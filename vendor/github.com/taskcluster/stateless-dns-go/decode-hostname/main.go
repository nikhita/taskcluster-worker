// decode-hostname is a command line tool for decoding stateless dns server names
package main

import (
	"fmt"
	"log"
	"os"

	docopt "github.com/docopt/docopt-go"
	"github.com/taskcluster/stateless-dns-go/hostname"
)

var (
	version = "decode-hostname 1.0.0"
	usage   = `
Usage:
  decode-hostname --fqdn FQDN --subdomain SUBDOMAIN --secret SECRET
  decode-hostname --help|-h
  decode-hostname --version

Exit Codes:
   0: Success
   1: Unrecognised command line options

Example:
  $ decode-hostname --fqdn 'znzsgaqaau2hl7h35f4owqn25s76j4h7apm3fe4qpy6pfxjk.foo.com' --subdomain 'foo.com' --secret 'cheese monkey'
`
)

func main() {

	arguments, err := docopt.Parse(usage, nil, true, version, false, true)
	if err != nil {
		fmt.Println("Error parsing command line arguments!")
		os.Exit(1)
	}

	fqdn := arguments["FQDN"].(string)
	subdomain := arguments["SUBDOMAIN"].(string)
	secret := arguments["SECRET"].(string)

	ip, expires, salt, err := hostname.Decode(fqdn, secret, subdomain)

	if err != nil {
		log.Fatalf("Error occured: %v", err)
	}

	log.Printf("IP: %v", ip)
	log.Printf("Expires: %v", expires)
	log.Printf("Salt: %#v", salt)
}
