package main

import (
	"fmt"
	"strings"
)

func init() {
	Lists = append(Lists,
		badips,
		bambenekconsulting,
		cinsscore,
		malwaredomainlist,
		malwaredomains,
		malc0de,
	)
}

var badips = func() List {
	fn := func(category, description string) func(row []string) *Entry {
		return func(row []string) *Entry {
			return &Entry{
				Source:      "badips.com",
				IP4:         row[0],
				Category:    category,
				Description: description,
			}
		}
	}

	return Combine(
		CSV("https://www.badips.com/get/list/ssh/3", fn("attacker", "SSH bruteforce login attacks and other ssh related attacks")),
		CSV("https://www.badips.com/get/list/dns/3", fn("attacker", "Attacks against the Domain Name System")),
		CSV("https://www.badips.com/get/list/http/3", fn("attacker", "Attacks aiming at HTTP/S services")),
	)
}()

var bambenekconsulting = CSV(
	"http://osint.bambenekconsulting.com/feeds/c2-ipmasterlist.txt",
	func(row []string) *Entry {
		return &Entry{
			Source:      "bambenekconsulting.com",
			IP4:         row[0],
			Category:    "malware",
			Description: row[1],
		}
	})

var cinsscore = CSV(
	"http://cinsscore.com/list/ci-badguys.txt",
	func(row []string) *Entry {
		return &Entry{
			Source:      "cinsscore.com",
			IP4:         row[0],
			Category:    "malware",
			Description: "",
		}
	})

var malwaredomainlist = CSV(
	"http://www.malwaredomainlist.com/mdlcsv.php",
	func(row []string) *Entry {
		return &Entry{
			Source:      "malwaredomainlist.com",
			Domain:      ExtractHost(row[1]),
			IP4:         ExtractHost(row[2]),
			Category:    "malware",
			Description: row[4],
		}
	})

var malwaredomainsCategories = map[string]string{
	"phishing":    "phishing",
	"malicious":   "malware",
	"attack_page": "attacker",
	"attackpage":  "attacker",
	"malware":     "malware",
	"botnet":      "botnet",
	"bedep":       "malware",
	"zeus":        "malware",
	"ransomware":  "malware",
	"malspam":     "malware",
	"simda":       "malware",
	"cryptowall":  "malware",
}

var malwaredomains = TSV(
	"http://mirror1.malwaredomains.com/files/domains.txt",
	func(row []string) *Entry {
		var e *Entry
		if strings.HasPrefix(row[0], "20") {
			// row has a next validation info
			e = &Entry{
				Source:      "malwaredomains.com",
				Domain:      row[1],
				Category:    row[2],
				Description: fmt.Sprintf("%q marked it as %q", row[3], row[2]),
			}
		} else {
			// row has no next validation info, this tricks the csv parser
			e = &Entry{
				Source:      "malwaredomains.com",
				Domain:      row[0],
				Category:    row[1],
				Description: fmt.Sprintf("%q marked it as %q", row[2], row[1]),
			}
		}

		if category, ok := malwaredomainsCategories[e.Category]; ok {
			e.Category = category
			return e
		}

		return nil
	})

var malc0de = Combine(
	SSV2("http://malc0de.com/bl/BOOT",
		func(row []string) *Entry {
			if len(row) < 2 {
				return nil
			}

			return &Entry{
				Source:      "malc0de.com",
				Domain:      row[1],
				Category:    "malware",
				Description: "distributed malware in the last 30 days",
			}
		}),
	CSV2("http://malc0de.com/bl/IP_Blacklist.txt",
		func(row []string) *Entry {
			if len(row) < 1 {
				return nil
			}

			return &Entry{
				Source:      "malc0de.com",
				IP4:         row[0],
				Category:    "malware",
				Description: "distributed malware in the last 30 days",
			}
		}))
