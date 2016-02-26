package main

func init() {
	Lists = append(Lists,
		//badips,
		bambenekconsulting,
		cinsscore,
		malwaredomainlist)
}

var badips = func() List {
	fn := func(category string) func(row []string) *Entry {
		return func(row []string) *Entry {
			return &Entry{
				Source:   "badips.com",
				Category: category,
				IP4:      row[0],
			}
		}
	}

	return Combine(
		CSV("https://www.badips.com/get/list/ssh/3", fn("attacker")),
		CSV("https://www.badips.com/get/list/cms/3", fn("attacker")),
		CSV("https://www.badips.com/get/list/http/3", fn("attacker")))
}()

var bambenekconsulting = CSV(
	"http://osint.bambenekconsulting.com/feeds/c2-ipmasterlist.txt",
	func(row []string) *Entry {
		return &Entry{
			Source:      "bambenekconsulting.com",
			Category:    "malware",
			IP4:         row[0],
			Description: row[1],
		}
	})

var cinsscore = CSV(
	"http://cinsscore.com/list/ci-badguys.txt",
	func(row []string) *Entry {
		return &Entry{
			Source:   "cinsscore.com",
			Category: "malware",
			IP4:      row[0],
		}
	})

var malwaredomainlist = CSV(
	"http://www.malwaredomainlist.com/mdlcsv.php",
	func(row []string) *Entry {
		return &Entry{
			Source:      "malwaredomainlist.com",
			Category:    "malware",
			Domain:      ExtractHost(row[1]),
			IP4:         ExtractHost(row[2]),
			Description: row[4],
		}
	})
