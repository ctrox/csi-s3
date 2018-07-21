package s3

import "fmt"

// Mounter interface which can be implemented
// by the different mounter types
type Mounter interface {
	Format() error
	Mount(targetPath string) error
	Unmount(targetPath string) error
}

const (
	s3fsMounterType   = "s3fs"
	goofysMounterType = "goofys"
	s3qlMounterType   = "s3ql"
)

// newMounter returns a new mounter depending on the mounterType parameter
func newMounter(bucket string, cfg *Config) (Mounter, error) {
	switch cfg.Mounter {
	case s3fsMounterType:
		return newS3fsMounter(bucket, cfg)

	case goofysMounterType:
		return newGoofysMounter(bucket, cfg)

	case s3qlMounterType:
		return newS3qlMounter(bucket, cfg)

	}
	return nil, fmt.Errorf("Error mounting bucket %s, invalid mounter specified: %s", bucket, cfg.Mounter)
}
