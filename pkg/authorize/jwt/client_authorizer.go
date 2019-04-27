package jwt

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"strings"
	"github.com/openshift/telemeter/pkg/authorize"
	"gopkg.in/square/go-jose.v2/jwt"
)

func NewClientAuthorizer(issuer string, keys []crypto.PublicKey, v Validator) *clientAuthorizer {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &clientAuthorizer{iss: issuer, keys: keys, validator: v}
}

type clientAuthorizer struct {
	iss		string
	keys		[]crypto.PublicKey
	validator	Validator
}

func (j *clientAuthorizer) AuthorizeClient(tokenData string) (*authorize.Client, bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if !j.hasCorrectIssuer(tokenData) {
		return nil, false, nil
	}
	tok, err := jwt.ParseSigned(tokenData)
	if err != nil {
		return nil, false, nil
	}
	public := &jwt.Claims{}
	private := j.validator.NewPrivateClaims()
	var (
		found	bool
		errs	[]error
	)
	for _, key := range j.keys {
		if err := tok.Claims(key, public, private); err != nil {
			errs = append(errs, err)
			continue
		}
		found = true
		break
	}
	if !found {
		return nil, false, multipleErrors(errs)
	}
	client, err := j.validator.Validate(tokenData, public, private)
	if err != nil {
		return nil, false, err
	}
	return client, true, nil
}
func (j *clientAuthorizer) hasCorrectIssuer(tokenData string) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	parts := strings.SplitN(tokenData, ".", 4)
	if len(parts) != 3 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	claims := struct {
		Issuer string `json:"iss"`
	}{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return false
	}
	if claims.Issuer != j.iss {
		return false
	}
	return true
}
