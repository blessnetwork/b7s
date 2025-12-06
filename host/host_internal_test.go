package host

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetermineAddressProtocol(t *testing.T) {

	tests := []struct {
		address   string
		shouldErr bool
		protocol  string
	}{
		{
			address:   "127.0.0.1",
			shouldErr: false,
			protocol:  "ip4",
		},
		{
			address:   "192.168.0.1",
			shouldErr: false,
			protocol:  "ip4",
		},
		{
			address:   "::FFFF:C0A8:1",
			shouldErr: false,
			protocol:  "ip6",
		},
		{
			address:   "0000:0000:0000:0000:0000:FFFF:C0A8:1",
			shouldErr: false,
			protocol:  "ip6",
		},
		{
			address:   "example.com",
			shouldErr: false,
			protocol:  "dns",
		},
		{
			address:   "foo1.bar2.com.gah.zip",
			shouldErr: false,
			protocol:  "dns",
		},
		{
			address:   "foo1.bar2.com.gah.zip",
			shouldErr: false,
			protocol:  "dns",
		},
		{
			address:   "hostname",
			shouldErr: false,
			protocol:  "dns",
		},
		{
			address:   "",
			shouldErr: true,
		},
		{
			address:   "698.168.0.1",
			shouldErr: true,
		},
		{
			// Documenting that we do NOT support certain things:
			//
			// While the Domain Name System (DNS) technically supports arbitrary sequences of octets in domain name labels,
			// the DNS standards recommend the use of the LDH (letter-digit-hyphen) subset of ASCII conventionally used for
			// host names, and require that string comparisons between DNS domain names should be case-insensitive.
			//
			// See => https://en.wikipedia.org/wiki/Punycode
			address:   "例子.測試",
			shouldErr: true,
		},
	}

	for _, test := range tests {
		protocol, address, err := determineAddressProtocol(test.address)
		if test.shouldErr {
			require.Error(t, err)
			break
		}

		require.NoError(t, err)
		require.Equal(t, test.address, address)
		require.Equalf(t, test.protocol, protocol, "unexpected protocol for address: %s", test.address)
	}
}
