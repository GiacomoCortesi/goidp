package controllers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

// ReadKeyPair parses the two raw keys (not encoded) into the corresponding data structures. The return
// value of the two key pointers to the keys is not to be relied upon if the returned error is not nil.
// The x509 parameter specifies if the passed public key is an RSA or PKIX (x509) key.
func ReadKeyPair(rawPubKey, rawPvtKey []byte, x509 bool) (*rsa.PublicKey, *rsa.PrivateKey, error) {
	if x509 {
		return readPKIXKeyPair(rawPubKey, rawPvtKey)
	} else {
		return readRSAKeyPair(rawPubKey, rawPvtKey)
	}
}

// IsX509Certificate returns true if the passed data contains a X509 certificate and false otherwise
func IsX509Certificate(data []byte) bool {
	_, err := x509.ParseCertificate(data)
	return err == nil
}

func readPrivateKey(input []byte) (*rsa.PrivateKey, error) {

	// Try PKCS1 first
	pvtKey, err := x509.ParsePKCS1PrivateKey(input)

	// Failure: fall-back to PKCS8 before giving in
	if err != nil {
		pkcs8, err := x509.ParsePKCS8PrivateKey(input)
		if err != nil {
			return nil, errors.New("private key error: neither PKCS1 nor PKCS8 format")
		}
		pvtKey, ok := (pkcs8).(*rsa.PrivateKey)
		if !ok {
			return nil,
				errors.New(fmt.Sprintf("private key error, expected rsa.PrivateKey, found '%T' instead", pvtKey))
		}
		return pvtKey, nil
	}
	return pvtKey, nil
}

func readRSAKeyPair(rawPubKey, rawPvtKey []byte) (*rsa.PublicKey, *rsa.PrivateKey, error) {
	pubKey, err := x509.ParsePKCS1PublicKey(rawPubKey)
	if err != nil {
		return nil, nil, errors.New("public key error " + err.Error())
	}
	pvtKey, err := readPrivateKey(rawPvtKey)
	if err != nil {
		return nil, nil, errors.New("private key error " + err.Error())
	}
	return pubKey, pvtKey, nil
}

func readPKIXKeyPair(rawCert, rawPvtKey []byte) (*rsa.PublicKey, *rsa.PrivateKey, error) {
	cert, err := x509.ParseCertificate(rawCert)
	if err != nil {
		return nil, nil, errors.New("public key error " + err.Error())
	}
	if pubKey, ok := cert.PublicKey.(*rsa.PublicKey); !ok {
		return nil, nil, errors.New("the certificate does not contain an RSA public key, found '" +
			fmt.Sprintf("%T", pubKey) +
			"' instead")
	}
	pvtKey, err := readPrivateKey(rawPvtKey)
	if err != nil {
		return nil, nil, errors.New("private key error " + err.Error())
	}
	return cert.PublicKey.(*rsa.PublicKey), pvtKey, nil
}

func ReadPublicKeys(publicKeyListPath string) []*rsa.PublicKey {
	var publicList []*rsa.PublicKey

	files, err := ioutil.ReadDir(publicKeyListPath)
	if err != nil {
		log.WithError(err).Warnf("failed to read keys dir")
		return nil
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		pubKeyPath := publicKeyListPath + "/" + file.Name()
		log.Infoln("currently reading: ", pubKeyPath)

		data, err := ioutil.ReadFile(pubKeyPath)
		if err != nil {
			log.WithError(err).Warnf("failed to read key data from file")
			continue
		}

		block, _ := pem.Decode(data)
		if block == nil {
			log.WithError(err).Warnf("invalid key")
			continue
		}

		pKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			log.WithError(err).Warnf("failed to parse key data")
			continue
		}
		rsaPKey := pKey.(*rsa.PublicKey)

		publicList = append(publicList, rsaPKey)
	}

	return publicList
}

// getIP is the utility function used to extract the ip from the request body
func getIP(r *http.Request) (string, error) {
	xForwardedFor := r.Header.Get("X-Forwarded-For")

	// X-Forwarded-For can be comma separated or comma and space separated
	xForwardedFor = strings.ReplaceAll(xForwardedFor, " ", "")
	xForwardedForIPs := strings.Split(xForwardedFor, ",")

	numForwardedIPs := len(xForwardedForIPs)
	if numForwardedIPs > 0 {
		realIP := xForwardedForIPs[numForwardedIPs-1]
		netRealIP := net.ParseIP(realIP)
		if netRealIP != nil {
			return realIP, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	return "", fmt.Errorf("no valid ip found")
}

func stringInSlice(sL []string, targetS string) bool {
	for _, s := range sL {
		if s == targetS {
			return true
		}
	}
	return false
}

func stringInSliceCaseInsensitive(sL []string, targetS string) bool {
	for _, s := range sL {
		if strings.ToLower(s) == strings.ToLower(targetS) {
			return true
		}
	}
	return false
}

func ReadPublicKey(publicKey string) (*rsa.PublicKey, error) {
	keyContent, err := ReadKey(publicKey)
	if err != nil {
		return nil, err
	}
	return x509.ParsePKCS1PublicKey(keyContent)
}

func ReadPrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	keyContent, err := ReadKey(privateKey)
	if err != nil {
		return nil, err
	}
	return x509.ParsePKCS1PrivateKey(keyContent)
}

func ReadKey(key string) ([]byte, error) {
	var pemString string
	if _, err := os.Stat(key); err == nil {
		// specified key is a valid file, read it
		data, err := ioutil.ReadFile(key)
		if err != nil {
			return nil, err
		}
		pemString = string(data)
	} else {
		pemString = key
	}

	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, errors.New("invalid key")
	}
	return block.Bytes, nil
}
