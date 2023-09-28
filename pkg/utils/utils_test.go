package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	i := 0
	errFormat := "`i is %d`"
	f1 := func() error {
		i++
		if i == 3 {
			return nil
		}
		return fmt.Errorf(errFormat, i)
	}

	testCases := []struct {
		testFunc       func() error
		retries        int
		i              int
		errShouldBeNil bool
		errorMsg       string
	}{
		{
			testFunc:       f1,
			retries:        0,
			i:              1,
			errShouldBeNil: false,
			errorMsg:       fmt.Sprintf(errFormat, 1),
		},
		{
			testFunc:       f1,
			retries:        1,
			i:              2,
			errShouldBeNil: false,
			errorMsg:       fmt.Sprintf(errFormat, 2),
		},
		{
			testFunc:       f1,
			retries:        2,
			i:              3,
			errShouldBeNil: true,
			errorMsg:       "",
		},
		{
			testFunc:       f1,
			retries:        3,
			i:              3,
			errShouldBeNil: true,
			errorMsg:       "",
		},
	}

	for j, tc := range testCases {
		i = 0
		t.Logf("Executing test case %d", j)
		err := ReTry(tc.testFunc, 1*time.Millisecond, tc.retries)
		if tc.errShouldBeNil && err != nil {
			t.Errorf("err shoul be nil, but is: %s", err.Error())
		} else if !tc.errShouldBeNil && err == nil {
			t.Errorf("err shoul not be %s, but is nil", tc.errorMsg)
		} else if !tc.errShouldBeNil && err != nil && err.Error() != tc.errorMsg {
			t.Errorf("error should be %s, but error message is: %s", tc.errorMsg, err.Error())
		}

		if i != tc.i {
			t.Errorf("i should be %d, but is: %d", tc.i, i)
		}
	}
}

func TestGetPullingImage(t *testing.T) {
	testCases := []struct {
		image   string
		message string
	}{
		{
			image:   "reg.docker.com/antsys/ulogfs-ilogtail-sidecar:dae94b0",
			message: `pulling image "reg.docker.com/antsys/ulogfs-ilogtail-sidecar:dae94b0"`,
		},
	}

	for _, tc := range testCases {
		s := GetPullingImageFromEventMessage(tc.message)
		if s != tc.image {
			t.Errorf("Image should be `%s`, but got `%s`", tc.image, s)
		}
	}
}
func TestGetPulledImage(t *testing.T) {
	testCases := []struct {
		image   string
		message string
	}{
		{
			image:   "reg.docker.com/antsys/ulogfs-sessmgr-sidecar:dae94b0",
			message: `Successfully pulled image "reg.docker.com/antsys/ulogfs-sessmgr-sidecar:dae94b0"`,
		},
		{
			image:   "reg.docker.com/aci/jenkins-slave-jnlp:2019-09-24",
			message: `Container image "reg.docker.com/aci/jenkins-slave-jnlp:2019-09-24" already present on machine`,
		},
		{
			image:   "reg.docker.com/antmesh/odp:1.3.7",
			message: `Successfully pulled image "reg.docker.com/antmesh/odp:1.3.7"`,
		},
	}

	for _, tc := range testCases {
		s := GetPulledImageFromEventMessage(tc.message)
		if s != tc.image {
			t.Errorf("Image should be `%s`, but got `%s`", tc.image, s)
		}
	}
}
