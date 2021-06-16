package middleware

import (
	"database/sql"
	"net"
	"net/http"

	"github.com/s-gv/orangeforum/models"
)

func IpFilter(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// go-chi's middleware/realip.go already parses for RealIp and xForwardedFor, further sets RemoteAddr to xForwardedFor ip if available.
		ipAddress, _, splitHostPortError := net.SplitHostPort(r.RemoteAddr)
		if splitHostPortError != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		parsedIp := net.ParseIP(ipAddress)
		if parsedIp == nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		isIpBanned, ipFilterCheckError := checkIfIpAddressIsBanned(parsedIp.String())

		if ipFilterCheckError != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if isIpBanned {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func checkIfIpAddressIsBanned(ipAddress string) (bool, error) {
	queriedIpFromDB, readError := getIpAddressFromDB(ipAddress)

	if readError != nil {
		return false, readError
	}

	return queriedIpFromDB != "", nil
}

func getIpAddressFromDB(ipAddressToBeQueried string) (string, error) {
	row := models.DB.QueryRow(`
								SELECT host(ip)
								FROM bannedips
								WHERE ip = $1`, ipAddressToBeQueried)

	var bannedIp string
	err := row.Scan(&bannedIp)

	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return bannedIp, nil
}
