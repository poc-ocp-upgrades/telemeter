package jwt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"fmt"
	"strings"
	"time"
	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

func NewSigner(issuer string, private crypto.PrivateKey) *Signer {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &Signer{iss: issuer, privateKey: private}
}

type Signer struct {
	iss		string
	privateKey	crypto.PrivateKey
}

func (j *Signer) GenerateToken(claims *jwt.Claims, privateClaims interface{}) (string, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var alg jose.SignatureAlgorithm
	switch privateKey := j.privateKey.(type) {
	case *rsa.PrivateKey:
		alg = jose.RS256
	case *ecdsa.PrivateKey:
		switch privateKey.Curve {
		case elliptic.P256():
			alg = jose.ES256
		case elliptic.P384():
			alg = jose.ES384
		case elliptic.P521():
			alg = jose.ES512
		default:
			return "", fmt.Errorf("unknown private key curve, must be 256, 384, or 521")
		}
	default:
		return "", fmt.Errorf("unknown private key type %T, must be *rsa.PrivateKey or *ecdsa.PrivateKey", j.privateKey)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: alg, Key: j.privateKey}, nil)
	if err != nil {
		return "", err
	}
	return jwt.Signed(signer).Claims(privateClaims).Claims(claims).Claims(&jwt.Claims{Issuer: j.iss}).CompactSerialize()
}
func multipleErrors(errs []error) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(errs) > 1 {
		return listErr(errs)
	}
	if len(errs) == 0 {
		return nil
	}
	return errs[0]
}

type listErr []error

func (errs listErr) Error() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var messages []string
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	return "multiple errors: " + strings.Join(messages, ", ")
}
func now() time.Time {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return time.Now()
}
