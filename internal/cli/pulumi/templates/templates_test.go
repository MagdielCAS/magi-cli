package templates

import (
	"strings"
	"testing"
)

func TestGetPulumiYamlTemplate(t *testing.T) {
	got := GetPulumiYamlTemplate("test-project", "test description")
	if !strings.Contains(got, "name: test-project") {
		t.Errorf("GetPulumiYamlTemplate() missing project name")
	}
	if !strings.Contains(got, "description: test description") {
		t.Errorf("GetPulumiYamlTemplate() missing description")
	}
}

func TestGetPackageJsonTemplate(t *testing.T) {
	got := GetPackageJsonTemplate("test-project")
	if !strings.Contains(got, "\"name\": \"test-project\"") {
		t.Errorf("GetPackageJsonTemplate() missing project name")
	}
	if !strings.Contains(got, "@pulumi/pulumi") {
		t.Errorf("GetPackageJsonTemplate() missing pulumi dependency")
	}
}

func TestGetTsConfigTemplate(t *testing.T) {
	got := GetTsConfigTemplate()
	if !strings.Contains(got, "\"target\": \"es2016\"") {
		t.Errorf("GetTsConfigTemplate() missing target")
	}
}

func TestAWS_Templates(t *testing.T) {
	tests := []struct {
		name     string
		template string
		check    string
	}{
		{"S3", GetS3BucketTemplate("my-bucket"), "aws.s3.Bucket"},
		{"VPC", GetVpcTemplate("my-vpc"), "awsx.ec2.Vpc"},
		{"EC2", GetEc2InstanceTemplate("my-server", "ami-123"), "aws.ec2.Instance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(tt.template, tt.check) {
				t.Errorf("%s template missing expected content %s", tt.name, tt.check)
			}
		})
	}
}

func TestGetAllTemplates(t *testing.T) {
	all := GetAllTemplates()
	if len(all) < 3 {
		t.Errorf("GetAllTemplates() returned too few templates")
	}
	if _, ok := all["s3_bucket"]; !ok {
		t.Errorf("GetAllTemplates() missing s3_bucket")
	}
}
