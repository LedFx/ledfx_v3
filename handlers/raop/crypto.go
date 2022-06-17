package raop

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	log "ledfx/logger"
	"net"
	"strings"
)

// this stuff is based on another go airtunes server: https://github.com/joelgibson/go-airplay/blob/master/airplay/auth.go

//from Shairport: https://github.com/abrasive/shairport/
const privateKey string = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA59dE8qLieItsH1WgjrcFRKj6eUWqi+bGLOX1HL3U3GhC/j0Qg90u3sG/1CUt
wC5vOYvfDmFI6oSFXi5ELabWJmT2dKHzBJKa3k9ok+8t9ucRqMd6DZHJ2YCCLlDRKSKv6kDqnw4U
wPdpOMXziC/AMj3Z/lUVX1G7WSHCAWKf1zNS1eLvqr+boEjXuBOitnZ/bDzPHrTOZz0Dew0uowxf
/+sG+NCK3eQJVxqcaJ/vEHKIVd2M+5qL71yJQ+87X6oV3eaYvt3zWZYD6z5vYTcrtij2VZ9Zmni/
UAaHqn9JdsBWLUEpVviYnhimNVvYFZeCXg/IdTQ+x4IRdiXNv5hEewIDAQABAoIBAQDl8Axy9XfW
BLmkzkEiqoSwF0PsmVrPzH9KsnwLGH+QZlvjWd8SWYGN7u1507HvhF5N3drJoVU3O14nDY4TFQAa
LlJ9VM35AApXaLyY1ERrN7u9ALKd2LUwYhM7Km539O4yUFYikE2nIPscEsA5ltpxOgUGCY7b7ez5
NtD6nL1ZKauw7aNXmVAvmJTcuPxWmoktF3gDJKK2wxZuNGcJE0uFQEG4Z3BrWP7yoNuSK3dii2jm
lpPHr0O/KnPQtzI3eguhe0TwUem/eYSdyzMyVx/YpwkzwtYL3sR5k0o9rKQLtvLzfAqdBxBurciz
aaA/L0HIgAmOit1GJA2saMxTVPNhAoGBAPfgv1oeZxgxmotiCcMXFEQEWflzhWYTsXrhUIuz5jFu
a39GLS99ZEErhLdrwj8rDDViRVJ5skOp9zFvlYAHs0xh92ji1E7V/ysnKBfsMrPkk5KSKPrnjndM
oPdevWnVkgJ5jxFuNgxkOLMuG9i53B4yMvDTCRiIPMQ++N2iLDaRAoGBAO9v//mU8eVkQaoANf0Z
oMjW8CN4xwWA2cSEIHkd9AfFkftuv8oyLDCG3ZAf0vrhrrtkrfa7ef+AUb69DNggq4mHQAYBp7L+
k5DKzJrKuO0r+R0YbY9pZD1+/g9dVt91d6LQNepUE/yY2PP5CNoFmjedpLHMOPFdVgqDzDFxU8hL
AoGBANDrr7xAJbqBjHVwIzQ4To9pb4BNeqDndk5Qe7fT3+/H1njGaC0/rXE0Qb7q5ySgnsCb3DvA
cJyRM9SJ7OKlGt0FMSdJD5KG0XPIpAVNwgpXXH5MDJg09KHeh0kXo+QA6viFBi21y340NonnEfdf
54PX4ZGS/Xac1UK+pLkBB+zRAoGAf0AY3H3qKS2lMEI4bzEFoHeK3G895pDaK3TFBVmD7fV0Zhov
17fegFPMwOII8MisYm9ZfT2Z0s5Ro3s5rkt+nvLAdfC/PYPKzTLalpGSwomSNYJcB9HNMlmhkGzc
1JnLYT4iyUyx6pcZBmCd8bD0iwY/FzcgNDaUmbX9+XDvRA0CgYEAkE7pIPlE71qvfJQgoA9em0gI
LAuE4Pu13aKiJnfft7hIjbK+5kyb3TysZvoyDnb3HOKvInK7vXbKuU4ISgxB2bB3HcYzQMGsz1qJ
2gG0N5hvJpzwwhbhXqFKA4zaaSrw622wDniAK5MlIE0tIAKKP4yxNGjoD2QYjhBGuhvkWKY=
-----END RSA PRIVATE KEY-----`

// these three functions are from the above git repo
func base64pad(s string) string {
	for len(s)%4 != 0 {
		s += "="
	}
	return s
}

func base64unpad(s string) string {
	if idx := strings.Index(s, "="); idx >= 0 {
		s = s[:idx]
	}
	return s
}

func aeskeyFromRsa(rsaaeskey64 string) (key []byte, err error) {
	s64 := base64pad(rsaaeskey64)
	s, err := base64.StdEncoding.DecodeString(s64)
	if err != nil {
		return
	}
	privKey, err := getPrivateKey()
	if err != nil {
		return nil, err
	}
	return rsa.DecryptOAEP(sha1.New(), nil, privKey, s, nil)
}

// the challenge response is the following
// 1. the base64 decoded data passed in as the challenge
// 2. the local connection IP address is added
// 3. the local connection's interface's mac address is added in (same that is used for the bonjour broadcast)
// 4. padding 0s are added if less than 32 bytes
// 5. the payload is signed with the private key
// 6. the signed data is base64 encoded
func generateChallengeResponse(challenge string, macAddr net.HardwareAddr, ipAddr string) (string, error) {

	log.Logger.WithField("context", "RAOP Crypto").Printf(fmt.Sprintf("building challenge for %s (ip: %s, mac: %s)", challenge, ipAddr, macAddr.String()))

	// the incoming challenge will be unpadded, need to pad to
	a := base64pad(challenge)
	decodedChallenge, err := base64.StdEncoding.DecodeString(a)
	if err != nil {
		return "", err
	}
	if len(decodedChallenge) != 16 {
		return "", fmt.Errorf("incorrect challenge received")
	}

	b := net.ParseIP(ipAddr)
	// ParseIP will always return a 16 byte array, so if we have a
	// ipv4 address we need to get the last 4 bytes only
	if b.To4() != nil {
		b = b[len(b)-4:]
	}

	decodedChallenge = append(decodedChallenge, b...)
	decodedChallenge = append(decodedChallenge, macAddr...)

	for len(decodedChallenge) < 32 {
		decodedChallenge = append(decodedChallenge, 0)
	}

	log.Logger.WithField("context", "RAOP Crypto").Println(hex.EncodeToString(decodedChallenge))

	rsaPrivKey, err := getPrivateKey()
	if err != nil {
		return "", err
	}

	signedResponse, err := rsa.SignPKCS1v15(nil, rsaPrivKey, crypto.Hash(0), decodedChallenge)
	if err != nil {
		return "", err
	}

	signedResponse64 := base64.StdEncoding.EncodeToString(signedResponse)

	if len(signedResponse64) != len(challenge) {
		signedResponse64 = base64unpad(signedResponse64)
	}

	log.Logger.WithField("context", "RAOP Crypto").Printf("Generated challenge response: %s\n", signedResponse64)
	return signedResponse64, nil
}

func getPrivateKey() (*rsa.PrivateKey, error) {
	pemBlock, _ := pem.Decode([]byte(privateKey))
	key, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}
