package s3_test

import (
	"log"
	"os"
	"path"

	"github.com/ctrox/csi-s3/pkg/s3"
	"github.com/kubernetes-csi/csi-test/pkg/sanity"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("S3Driver", func() {

	Context("goofys", func() {
		socket := "/tmp/csi-goofys.sock"
		csiEndpoint := "unix://" + socket
		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		driver, err := s3.NewS3("test-node", csiEndpoint)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  os.TempDir() + "/goofys-target",
				StagingPath: os.TempDir() + "/goofys-staging",
				Address:     csiEndpoint,
				SecretsFile: "../../test/secret.yaml",
				TestVolumeParameters: map[string]string{
					"mounter": "goofys",
				},
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})

	Context("s3fs", func() {
		socket := "/tmp/csi-s3fs.sock"
		csiEndpoint := "unix://" + socket
		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		driver, err := s3.NewS3("test-node", csiEndpoint)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  os.TempDir() + "/s3fs-target",
				StagingPath: os.TempDir() + "/s3fs-staging",
				Address:     csiEndpoint,
				SecretsFile: "../../test/secret.yaml",
				TestVolumeParameters: map[string]string{
					"mounter": "s3fs",
				},
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})

	Context("s3backer", func() {
		socket := "/tmp/csi-s3backer.sock"
		csiEndpoint := "unix://" + socket

		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		// Clear loop device so we cover the creation of it
		os.Remove(s3.S3backerLoopDevice)
		driver, err := s3.NewS3("test-node", csiEndpoint)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  os.TempDir() + "/s3backer-target",
				StagingPath: os.TempDir() + "/s3backer-staging",
				Address:     csiEndpoint,
				SecretsFile: "../../test/secret.yaml",
				TestVolumeParameters: map[string]string{
					"mounter": "s3backer",
				},
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})

	Context("rclone", func() {
		socket := "/tmp/csi-rclone.sock"
		csiEndpoint := "unix://" + socket

		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		driver, err := s3.NewS3("test-node", csiEndpoint)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  os.TempDir() + "/rclone-target",
				StagingPath: os.TempDir() + "/rclone-staging",
				Address:     csiEndpoint,
				SecretsFile: "../../test/secret.yaml",
				TestVolumeParameters: map[string]string{
					"mounter": "rclone",
				},
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})
})

// var _ = Describe("S3Driver", func() {
// 	for _, mounter := range []string{"s3fs"} {
// 		// for _, mounter := range []string{"goofys", "s3fs", "s3backer", "rclone"} {
// 		Context(mounter, func() {
// 			for _, commonBucket := range []string{"", "csi-bucket"} {
// 				for _, commonPrefix := range []string{"", "csi/prefix"} {

// 					label := strings.Join([]string{mounter, commonBucket, strings.ReplaceAll(commonPrefix, "/", "-")}, "_")

// 					Describe("CSI sanity: "+label, func() {
// 						socket := fmt.Sprintf("/tmp/csi-%s.sock", label)
// 						csiEndpoint := "unix://" + socket
// 						It("creates driver socket", func() {
// 							if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
// 								Expect(err).NotTo(HaveOccurred())
// 							}
// 						})

// 						It("starts the driver", func() {
// 							driver, err := s3.NewS3("test-node", csiEndpoint)
// 							Expect(err).To(Succeed())
// 							go driver.Run()
// 						})

// 						secret := newSecret()
// 						for _, v := range secret {
// 							v["mounter"] = mounter
// 							v["commonPrefix"] = commonPrefix
// 							v["commonBucket"] = commonBucket
// 						}

// 						secretFile := path.Join(os.TempDir(), label, "secret.yaml")
// 						It("creates the secret file", func() {
// 							Expect(writeSecret(secret, secretFile)).NotTo(HaveOccurred())
// 						})

// 						sanity.GinkgoTest(&sanity.Config{
// 							TargetPath:  fmt.Sprintf("%s/%s-target", os.TempDir(), label),
// 							StagingPath: fmt.Sprintf("%s/%s-staging", os.TempDir(), label),
// 							Address:     csiEndpoint,
// 							SecretsFile: secretFile,
// 							TestVolumeParameters: map[string]string{
// 								"mounter": mounter,
// 							},
// 						})
// 					})
// 				}
// 			}
// 		})
// 	}
// })

func writeSecret(secret map[string]map[string]string, filename string) error {
	if e := os.MkdirAll(path.Dir(filename), os.ModeDir); e != nil {
		return e
	}
	bytes, e := yaml.Marshal(secret)
	if e != nil {
		return e
	}
	f, e := os.Create(filename)
	if e != nil {
		return e
	}
	defer f.Close()

	_, e = f.Write(bytes)
	return e
}

// deep copy and return a secret
func newSecret() map[string]map[string]string {
	secret := make(map[string]map[string]string, len(commonSecret))
	for k, v := range commonSecret {
		innerMap := make(map[string]string, len(v))
		for innerKey, innerValue := range v {
			innerMap[innerKey] = innerValue
		}
		secret[k] = innerMap
	}
	return secret
}

var commonSecret = map[string]map[string]string{
	"CreateVolumeSecret": {
		"accessKeyID":     "FJDSJ",
		"secretAccessKey": "DSG643HGDS",
		"endpoint":        "http://127.0.0.1:9000",
		"region":          "",
	},
	"DeleteVolumeSecret": {
		"accessKeyID":     "FJDSJ",
		"secretAccessKey": "DSG643HGDS",
		"endpoint":        "http://127.0.0.1:9000",
		"region":          "",
	},
	"NodeStageVolumeSecret": {
		"accessKeyID":     "FJDSJ",
		"secretAccessKey": "DSG643HGDS",
		"endpoint":        "http://127.0.0.1:9000",
		"region":          "",
	},
	"NodePublishVolumeSecret": {
		"accessKeyID":     "FJDSJ",
		"secretAccessKey": "DSG643HGDS",
		"endpoint":        "http://127.0.0.1:9000",
		"region":          "",
	},
	"ControllerValidateVolumeCapabilitiesSecret": {
		"accessKeyID":     "FJDSJ",
		"secretAccessKey": "DSG643HGDS",
		"endpoint":        "http://127.0.0.1:9000",
		"region":          "",
	},
}
