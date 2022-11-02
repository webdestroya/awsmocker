package awsmocker_test

import (
	"testing"
	"time"

	"github.com/webdestroya/awsmocker"
)

func TestCheckCertificateExpiry(t *testing.T) {
	cert := awsmocker.CACert()

	if cert.NotAfter.Before(time.Now().AddDate(3, 0, 0)) {
		t.Fatal("CA Certificate expires within 3 years. Should be regenerated")
	}

}
