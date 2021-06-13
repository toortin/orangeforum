package orangemiddleware

import (
	"net"
	"net/http"
	"sort"
)

func IpFilter(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// go-chi's middleware/realip.go already parses for RealIp and xForwardedFor, further sets RemoteAddr to xForwardedFor ip if available.
		ipAddress, _, splitHostPortError := net.SplitHostPort(r.RemoteAddr)
		if splitHostPortError != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		parsedIp := net.ParseIP(ipAddress)
		if parsedIp == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if checkIfIpIsBlocked(parsedIp.String()) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func checkIfIpIsBlocked(ipAddress string) bool {
	ipAddreses := getMockIpAddresses()
	return findIpAddress(ipAddreses, ipAddress)
}

//TO-DO look for a better algorithm
func findIpAddress(ipAddresses []string, ipAddress string) bool {
	sort.Strings(ipAddresses)
	searchResIndex := sort.SearchStrings(ipAddresses, ipAddress)
	return searchResIndex < len(ipAddresses) && ipAddresses[searchResIndex] == ipAddress
}

func getMockIpAddresses() []string {
	return []string{"192.168.1.120", "192.167.1.2", "0.0.0.0", "56.78.123.45"}
}
