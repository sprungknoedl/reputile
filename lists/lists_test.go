package lists

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

func init() {
	logrus.SetLevel(logrus.WarnLevel)
}

func TestBadIPs(t *testing.T) {
	test(t, badips)
}

func TestBambenekConsulting(t *testing.T) {
	test(t, bambenekconsulting)
}

func TestCinsScore(t *testing.T) {
	test(t, cinsscore)
}

func TestMalwareDomainList(t *testing.T) {
	test(t, malwaredomainlist)
}

func TestMalwareDomains(t *testing.T) {
	test(t, malwaredomains)
}

func TestMalc0de(t *testing.T) {
	test(t, malc0de)
}

func TestAutoshun(t *testing.T) {
	test(t, autoshun)
}

func TestBlocklist(t *testing.T) {
	test(t, blocklist)
}

func TestBruteforceblocker(t *testing.T) {
	test(t, bruteforceblocker)
}

func TestPhishtank(t *testing.T) {
	test(t, phishtank)
}

func TestAbuseCh(t *testing.T) {
	test(t, abusech)
}

func test(t *testing.T, list List) {
	t.Parallel()
	ctx := context.Background()
	count := 0

	for entry := range list(ctx) {
		if entry.Err != nil {
			t.Fatal(entry.Err)
		}

		count++
	}

	t.Logf("produced %d entries", count)
	if count == 0 {
		t.Fail()
	}
}
