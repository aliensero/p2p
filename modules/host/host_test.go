package host

import "testing"

func TestBootstrapInfo(t *testing.T) {
	pis, err := getBootstrapAddrInfos()
	t.Logf("pis %v error %v", pis, err)
}
