package s3

// Credentials holds s3 credentials and parameters
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Endpoint        string
}
