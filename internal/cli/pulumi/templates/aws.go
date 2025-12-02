package templates

import "fmt"

// Common AWS resource snippets that can be injected into the context
const (
	AWSProviderImport  = `import * as aws from "@pulumi/aws";`
	AWSXProviderImport = `import * as awsx from "@pulumi/awsx";`
)

// GetS3BucketTemplate returns a basic S3 bucket definition
func GetS3BucketTemplate(name string) string {
	return fmt.Sprintf(`const %sBucket = new aws.s3.Bucket("%s", {
    acl: "private",
    versioning: {
        enabled: true,
    },
});`, name, name)
}

// GetVpcTemplate returns a basic VPC definition using awsx
func GetVpcTemplate(name string) string {
	return fmt.Sprintf(`const %sVpc = new awsx.ec2.Vpc("%s", {
    cidrBlock: "10.0.0.0/16",
    numberOfAvailabilityZones: 2,
    subnetSpecs: [
        { type: awsx.ec2.SubnetType.Public, cidrMask: 24 },
        { type: awsx.ec2.SubnetType.Private, cidrMask: 24 },
    ],
});`, name, name)
}

// GetEc2InstanceTemplate returns a basic EC2 instance definition
func GetEc2InstanceTemplate(name, amiId string) string {
	return fmt.Sprintf(`const %sServer = new aws.ec2.Instance("%s", {
    instanceType: "t3.micro",
    ami: "%s", // e.g., ami-0c55b159cbfafe1f0
    tags: {
        Name: "%s",
    },
});`, name, name, amiId, name)
}

// GetAllTemplates returns all available AWS templates as a map
func GetAllTemplates() map[string]string {
	return map[string]string{
		"s3_bucket":    GetS3BucketTemplate("my-bucket"),
		"vpc":          GetVpcTemplate("my-vpc"),
		"ec2_instance": GetEc2InstanceTemplate("my-server", "ami-12345678"),
	}
}
