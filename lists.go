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
		phishtank,
		autoshun,
		blocklist,
		bruteforceblocker,
		abusech,
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
				Description: fmt.Sprintf("%s marked it as %s", row[3], row[2]),
			}
		} else {
			// row has no next validation info, this tricks the csv parser
			e = &Entry{
				Source:      "malwaredomains.com",
				Domain:      row[0],
				Category:    row[1],
				Description: fmt.Sprintf("%s marked it as %s", row[2], row[1]),
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

var phishtank = CSV(
	"http://data.phishtank.com/data/online-valid.csv",
	func(row []string) *Entry {
		if row[0] == "phish_id" {
			// header line
			return nil
		}

		domain := ExtractHost(row[1])
		if domain == "" {
			return nil
		}

		return &Entry{
			Source:      "phishtank.com",
			Domain:      domain,
			Category:    "phishing",
			Description: "Domain hosts web pages used for phishing",
		}
	},
)

var autoshun = CSV(
	"https://www.autoshun.org/files/shunlist.csv",
	func(row []string) *Entry {
		if len(row) < 3 {
			return nil
		}

		return &Entry{
			Source:      "autoshun.org",
			IP4:         row[0],
			Category:    "attacker",
			Description: row[2],
		}
	},
)

var blocklist = CSV(
	"http://lists.blocklist.de/lists/all.txt",
	func(row []string) *Entry {
		return &Entry{
			Source:   "blocklist.de",
			IP4:      row[0],
			Category: "attacker",
		}
	},
)

var bruteforceblocker = TSV(
	"http://danger.rulez.sk/projects/bruteforceblocker/blist.php",
	func(row []string) *Entry {
		return &Entry{
			Source:   "rulez.sk",
			IP4:      row[0],
			Category: "attacker",
		}
	},
)

var abusech = func() List {
	domain := func(description string) func(row []string) *Entry {
		return func(row []string) *Entry {
			return &Entry{
				Source:      "abuse.ch",
				Domain:      row[0],
				Category:    "malware",
				Description: description,
			}
		}
	}

	ip := func(description string) func(row []string) *Entry {
		return func(row []string) *Entry {
			return &Entry{
				Source:      "abuse.ch",
				IP4:         row[0],
				Category:    "malware",
				Description: description,
			}
		}
	}

	return Combine(
		CSV("https://zeustracker.abuse.ch/blocklist.php?download=baddomains", domain("ZeuS C&C server")),
		CSV("https://zeustracker.abuse.ch/blocklist.php?download=badips", ip("ZeuS C&C server")),
		CSV("https://feodotracker.abuse.ch/blocklist/?download=domainblocklist", domain("Feodo trojan C&C server")),
		CSV("https://feodotracker.abuse.ch/blocklist/?download=badips", ip("Feodo trojan C&C server")),
		CSV("https://ransomwaretracker.abuse.ch/downloads/RW_DOMBL.txt", domain("Ransomware botnet C&C traffic")),
		CSV("https://ransomwaretracker.abuse.ch/downloads/RW_IPBL.txt", ip("Ransomware botnet C&C traffic")),
	)
}()
