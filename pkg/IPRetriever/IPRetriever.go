package ipretriever

import (
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

type Provider string

const (
	ProviderAWS     Provider = "checkip.amazonaws.com"
	ProviderICanHaz Provider = "icanhazip.com"
)

type IpRetriever func() (string, error)

func AWS() (ip string, err error) {
	return getPublicIP(ProviderAWS)
}

func ICanHazIP() (ip string, err error) {
	return getPublicIP(ProviderICanHaz)
}

func getPublicIP(provider Provider) (ip string, err error) {
	resp, err := http.Get("https://" + string(provider))
	if err != nil {
		log.Err(err).Msg("Could not get public IP")
		return
	}
	defer resp.Body.Close()

	tokenizer := html.NewTokenizer(resp.Body)
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			err = tokenizer.Err()
			if err != io.EOF {
				log.Err(err).Msg("Error tokenizing HTML")
			}
			return
		case html.TextToken:
			ip = tokenizer.Token().String()
			return
		}
	}
}
