package runner

import (
	"log"
	"os/exec"

	ginkgoTestcase "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/testcase"
)

func GetGinkgoVersion(testcases []*ginkgoTestcase.TestCase) string {
	version := "2"
	for _, tc := range testcases {
		version = tc.Attributes["ginkgoVersion"]
	}
	log.Printf("ginkgo version is %s", version)
	return version
}

func CheckGinkgoCli() bool {
	if _, err := exec.LookPath("ginkgo"); err != nil {
		log.Println("ginkgo client not found")
		return false
	}
	return true
}
