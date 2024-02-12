package controllers

import (
	"encoding/pem"
	"testing"
)

// openssl rsa -pubin -in public.pem -text
// openssl rsa -in private.pem -text
// openssl x509 -in certificate.pem -text
// openssl rsa -pubin -in public.PKIX.pem -RSAPublicKey_out -out public.RSA.pem
// The following keys are PEM encoded:
// 1. X509 certificate
// 2. RSA private key matching (1)
// 3. Public matching (1) in PKIX format
// 4. RSA public key matching (1)
const X509_Certificate = `-----BEGIN CERTIFICATE-----
MIIDAjCCAeqgAwIBAgIRAI8PjyF+k7ckCSb/IBXMpNswDQYJKoZIhvcNAQELBQAw
HDEaMBgGA1UEAxMRb3J1Yy1pZHAuZnVsbG5hbWUwHhcNMjMwNzEwMTQ1ODM0WhcN
MzMwNzA3MTQ1ODM0WjAcMRowGAYDVQQDExFvcnVjLWlkcC5mdWxsbmFtZTCCASIw
DQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAOgYQKp6S1nLk4YtlMVzfAep/8/M
dRhs1WCSaygpFX7cnYhE1FqTDJdZSTpyLwZQxbK7j4qyCsVrNmWdeYCKn4TNou6t
JY4Er5ALCA1VudsJrc0dqdu1m0snsJPiLfE+u/vHhf+HVlJgvJkUVIbUuhcJKxnP
uGEx/OMy0lQinpJysCYCoPhJXR7+pWEw90RcQ/t+G+GGRGWhDY07xgGwclvfJ1Cr
OvlXozJEwuiZ3c0u5iVRI+3ebkWrcg2pRfQmoxVBXwUptId4K12cGqfcpZpvjalh
N68LbBLqFh+cD3xjJlGsdGf8ciN2XtnuJMQG//89oy7h32HPvcWyuq6ZScECAwEA
AaM/MD0wDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEF
BQcDAjAMBgNVHRMBAf8EAjAAMA0GCSqGSIb3DQEBCwUAA4IBAQAFphZt3+jBDC4+
0r01mMCj03Xsnjl6xa59JDybmH0JDdqwuzv5O8Q1wDExUAmXXQSUJm71Wi8/yagC
gygxV889pEdv8p2vIl/ZS+DpzDdHXz8l3BhhR18wo23s04RGg6z3PE/CLjdMhp7Q
Lq5zEr2D2ivigE+RKBQp0y7A6OadnteUp5Ju7dweJJDGjWJ9GxYzSFoRSBOPWKC6
sRRuzvGL1f+rtzSOeHQw1LHKCWqI+PEmqDd6ZyZ5pXbrDGHAw1LkUypl/5Mo572I
TREYFQHKLFQR/W6pDsQvEuEn3/80y9UZrOLibZEcHuuy5ASu/gR7zYuMBdUqiOps
GKYhFRAr
-----END CERTIFICATE-----`

// This is a PKCS1
const PKCS1_Private_Key = `-----BEGIN RSA PRIVATE KEY-----
MIIEpgIBAAKCAQEA6BhAqnpLWcuThi2UxXN8B6n/z8x1GGzVYJJrKCkVftydiETU
WpMMl1lJOnIvBlDFsruPirIKxWs2ZZ15gIqfhM2i7q0ljgSvkAsIDVW52wmtzR2p
27WbSyewk+It8T67+8eF/4dWUmC8mRRUhtS6FwkrGc+4YTH84zLSVCKeknKwJgKg
+EldHv6lYTD3RFxD+34b4YZEZaENjTvGAbByW98nUKs6+VejMkTC6JndzS7mJVEj
7d5uRatyDalF9CajFUFfBSm0h3grXZwap9ylmm+NqWE3rwtsEuoWH5wPfGMmUax0
Z/xyI3Ze2e4kxAb//z2jLuHfYc+9xbK6rplJwQIDAQABAoIBAQCM5VQ0acNegrhP
B1K+Pyo3WNtD4cHgDwnF83z7x10WQ4WamPY0+fn10y0iPvkPI2+w3i34q7bgPAKs
01lUUFMggtl7fT9EJNITZq7/sV//ebO5xl08VNYuXKzUScVMI6Jo6aoOArHDlphH
cdESfQdvPpCcvb4XuwnjPxHyI4YSLkd9bmPa1gbnls9a1bzqRFUGOGWQ68n/XG6E
R5sdfJZpA9tf24KoGMiBKLcGVCo6LcxWTwDz+OrAx5l2jFIpYboLmMY8M0jmbsjj
o5z5CT1UgYVsvyMzMWTayl5kUQEDiYf94nRf0GOvlCBfRdfVPbsfg/Qn2a7pXx/9
GwpXi73BAoGBAPkU1BBH0jTEmXa075NaP6nW2jzrG37eDdppTGLZ0/R8IMlYtW8M
NEPDRG+MBuydbYb9EDVGMBzjFIPpcdwR9vjy8wfNSJOhkeQKGIs2NThhzx5QKVMP
cz+CEUJhQFPWn6v1L/NCk+/1ybcO87XykrN5Cr7ZavrGqxHt2Y66MIp5AoGBAO6K
oq6QSxUXJjNxdKFo6PtkU7KHmtNbRsvZCsvOpc4IPb5d8VCXOINR1PY22m1DMAlt
tnpCiv5pv75GDQfmSsvmj+JsnSpvmGLJS7eQ3EI4co6Kqh0d8KzRnBISLdvUb8K3
LCrA7wbYz1Ts+9DNswVbKe5ANKGTm602VWbADOeJAoGBAMqt+CGHT7VAhN/jO09c
EJHTEqKfbTA+4GbpaA7H0YEPwF4WoQxLkfvR2M4r0zaWo5lEMvwmsN/Qp9DvFIdO
1vicOMYQdQ4sWtqEnJQq+AN5E2BHOlksKUt0OzcYi4+tBSCX0vzPIDISfqFGuWlE
ibsgs2243SRSpMFiGbXaK8WBAoGBAIrPoBWHIDoYq4E6H42iGBnaax4z7TPbJNqQ
5cht62x2vT9fOYMVTKyWXSAeEYONmpGSB6Mjv/CGpy7ZYtHbAGGhPM+dNuQv5nRu
ASLEKHhcksVCCfZBqwFWRMT7UTZga9zabNhAR5graJNaCLucR/Xw8/iR4k64L9pf
CNlvHtNJAoGBAPGPMaAJO3DfwslSy7NoLlhRfUeX8g+LbK0x4aVKGe7DkT9Rdl5/
iPTSsVebQdqzhTyVHJ8VMmu75Y1/pC4uJDHHJ4eYGF4PrvwlmyErj233KRCJvGN9
pt+ybJuCHiPK8wDTfeiSBWS9tZjzD7zh+PVL9y95mf2mcEplx6CRM0dJ
-----END RSA PRIVATE KEY-----`

const PKCS8_Private_Key = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQCxwmFM8n2DBBKs
urdn9hH/s85GgsR/7Oa4SgTZ3BB1Zql8S3Bjk4SyjuoynqO17j06czknJldLdM/D
TDXgHId8fyRD2V+Na1Zqj/CRswpD55yYHIqRvDsqPuFVfBmafPafmWuIEufZ6VxW
jGFgRlQzEhUPpVBKRGfCk/WhJPTWtl9/cuv9MxvpoZpm1Tv6Ln5z7xgmUXVpVuzl
R8R3zuRwknumCGR8NBDRkttR0BOI4Qk5Ur9qgDQMBEC78ghhr03lS6de1s8bimsc
VVvX0GoDwqC/XlyD4hx1DpBl4KgnQ+3ojK4Md+qdIZdlivdWafRLGR9yzR9U/KSE
Rc2DtxvVAgMBAAECggEAGLWifOXAWLP6PJR/5i2odtjxtY977SRrNfbkEbyrdQZe
TO7Xop/g9Ek1eO/gZevGCxf1O+HyhISqVMWFP6/3jXDHA791rtza2FlF4Zr3tFS/
yc093eBCgS7Yd9+WV4lDZxAWiIXIQNxVf5pn9tAP6EF9N/2M2aYEnmGe5VWTyzy3
WsnNMTrZUpyLeivAfwct3zr4D/Z8oT8S0mOIdLUgLbwluBbJgIhDLXFjR6gZ0sCq
d8UYEmQJZGXo2MITRuc9/beT3ppY8otHpEVnIL1MJ26jrnCFmE4XfMNoeNU/o+7F
jaafnaPiwfgUMOE9hKBPtiBaCmOwzjvBanw/B+gYQQKBgQDaerEmyfW5PbmWog9z
5L9/kh20ewGQUELcauXoh91+HSgrrNY2qekjHul0USjcfVGVU/VeiLbeuD/Ot+3g
iOuhtKyOJ0b4ZyiEdrwqpIzhJolA7Id0x8n64qL366pe8uVn9IFM7NIWyNU3Xow/
eNhevjBl8JjOea54wpaVyMV3SQKBgQDQSXWd/5bJiUokEI30SvJtQFN16JzuYPgS
6OvsdDKudwnF0oPLtkEoUyifkY8s2T7v6+WBC0Kt8g0/hw2nHeDQ+asGYmG1of98
cBCVAs5FrqqlbXIFjgR3XkvAJjKs/G/KrSs11zepEExtQ5jBtGCfbf8IOgxvYGA4
u4MAFgQELQKBgQCKiMnT4rPhJfaMQW6y+hVDew9C5cx0CbCbu1zVOXGFCk/ygcHD
H7IpBuzZSK00QnJ80aQAsYfjaclr9szrV2ayPrI74UPrNt5GQFPIZla+XYUimdi6
gATfBN55fgGl+zbj1/I1KOV+dRJd7aHYjXQFf2uI+CqsohOzlw+NIqWzoQKBgQCY
s3JuXkZ/BI79d8GKuzOWUxWdGOeMgDz/KBJm7R2G+LCKfnavb7O/S5A5xC5SdAcH
QEum2smM2ytJSssAnRAIRTJUYOY/kj/LTCFsDX2Kaq6iz2VLmz29Ab3JZne6iOuw
jFpkg59D7DYL2QXx5Tr9R1g1ANHDCcYhcZ9t/bX+FQKBgQC2CUYbZu/1ZtpD5lky
8nUX7N/T6FwGf/axJlzKzE9uO6izu8/Ka39x8YFZ8e9S96q7PAkeme+BVqhAlSDu
nd79sTntscfayOrOfEz+xrygFQ21mJhQGWRk9qEfpQGvu7+8kLvirbmadN3bvMNA
3EISy1QDlYBa9Q7Tdu8HEOsglg==
-----END PRIVATE KEY-----`

const PKIX_Public_Key = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA6BhAqnpLWcuThi2UxXN8
B6n/z8x1GGzVYJJrKCkVftydiETUWpMMl1lJOnIvBlDFsruPirIKxWs2ZZ15gIqf
hM2i7q0ljgSvkAsIDVW52wmtzR2p27WbSyewk+It8T67+8eF/4dWUmC8mRRUhtS6
FwkrGc+4YTH84zLSVCKeknKwJgKg+EldHv6lYTD3RFxD+34b4YZEZaENjTvGAbBy
W98nUKs6+VejMkTC6JndzS7mJVEj7d5uRatyDalF9CajFUFfBSm0h3grXZwap9yl
mm+NqWE3rwtsEuoWH5wPfGMmUax0Z/xyI3Ze2e4kxAb//z2jLuHfYc+9xbK6rplJ
wQIDAQAB
-----END PUBLIC KEY-----`

const RSA_Public_key = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA6BhAqnpLWcuThi2UxXN8B6n/z8x1GGzVYJJrKCkVftydiETUWpMM
l1lJOnIvBlDFsruPirIKxWs2ZZ15gIqfhM2i7q0ljgSvkAsIDVW52wmtzR2p27Wb
Syewk+It8T67+8eF/4dWUmC8mRRUhtS6FwkrGc+4YTH84zLSVCKeknKwJgKg+Eld
Hv6lYTD3RFxD+34b4YZEZaENjTvGAbByW98nUKs6+VejMkTC6JndzS7mJVEj7d5u
RatyDalF9CajFUFfBSm0h3grXZwap9ylmm+NqWE3rwtsEuoWH5wPfGMmUax0Z/xy
I3Ze2e4kxAb//z2jLuHfYc+9xbK6rplJwQIDAQAB
-----END RSA PUBLIC KEY-----`

func decodePem(data string) []byte {
	decoded, _ := pem.Decode([]byte(data))
	if decoded == nil {
		panic("Unable to decode data '" + data + "'")
	}
	return decoded.Bytes
}

func TestIsX509Certificate(t *testing.T) {
	if IsX509Certificate(decodePem(PKCS8_Private_Key)) {
		t.Error("The public key is in RSA format")
	}
	if !IsX509Certificate(decodePem(X509_Certificate)) {
		t.Error("The input is X509 certificate")
	}
}

func TestReadPrivateKeyPKCS1(t *testing.T) {
	_, err := readPrivateKey(decodePem(PKCS1_Private_Key))
	if err != nil {
		t.Error(err.Error())
	}
}

func TestReadPrivateKeyPKCS8(t *testing.T) {
	_, err := readPrivateKey(decodePem(PKCS8_Private_Key))
	if err != nil {
		t.Error(err.Error())
	}
}

func TestReadPrivateKeyNegative(t *testing.T) {
	_, err := readPrivateKey(decodePem(PKIX_Public_Key))
	if err == nil {
		t.Error("This is a negative test")
	}
}

func TestReadRSAKeyPairPositive(t *testing.T) {
	pubKey, pvtKey, err := readRSAKeyPair(decodePem(RSA_Public_key), decodePem(PKCS8_Private_Key))
	if err != nil {
		t.Error("errors found " + err.Error())
	}
	if pubKey == nil {
		t.Error("error: public key nil ")
	}
	if pvtKey == nil {
		t.Error("error: private key nil ")
	}
}

func TestReadRSAKeyPairNegative(t *testing.T) {
	_, _, err := readRSAKeyPair(decodePem(X509_Certificate), decodePem(PKCS8_Private_Key))
	if err == nil {
		t.Error("errors NOT found the test should fail")
	}
}

func TestReadPKIXKeyPairNegative(t *testing.T) {
	_, _, err := readPKIXKeyPair(decodePem(RSA_Public_key), decodePem(PKCS8_Private_Key))
	if err == nil {
		t.Error("errors not found, the test should fail " + err.Error())
	}
}

func TestReadPKIXKeyPairPositive(t *testing.T) {
	pubKey, pvtKey, err := readPKIXKeyPair(decodePem(X509_Certificate), decodePem(PKCS8_Private_Key))
	if err != nil {
		t.Error("errors found " + err.Error())
	}
	if pubKey == nil {
		t.Error("error: public key nil ")
	}
	if pvtKey == nil {
		t.Error("error: private key nil ")
	}
}

func TestReadKeyPairX509(t *testing.T) {
	pubKey, pvtKey, err := ReadKeyPair(decodePem(X509_Certificate), decodePem(PKCS8_Private_Key), true)
	if err != nil {
		t.Error("errors found " + err.Error())
	}
	if pubKey == nil {
		t.Error("error: public key nil ")
	}
	if pvtKey == nil {
		t.Error("error: private key nil ")
	}
}

func TestReadKeyPairNoX509(t *testing.T) {
	pubKey, pvtKey, err := ReadKeyPair(decodePem(RSA_Public_key), decodePem(PKCS8_Private_Key), false)
	if err != nil {
		t.Error("errors found " + err.Error())
	}
	if pubKey == nil {
		t.Error("error: public key nil ")
	}
	if pvtKey == nil {
		t.Error("error: private key nil ")
	}
}
