package agents

import (
	"reflect"
	"sort"
	"testing"
)

func TestInfrastructureValidator_extractResourceTypes(t *testing.T) {
	// We can't easily instantiate the full agent without mocks,
	// but we can test the helper method if we make it accessible or test via a dummy struct if it was an interface.
	// Since it's a method on the struct, we'll create a nil-safe wrapper or just instantiate the struct with nil dependencies
	// as long as we don't call methods that use them.

	v := &InfrastructureValidator{}

	tests := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "Single resource",
			code: `const bucket = new aws.s3.Bucket("my-bucket");`,
			want: []string{"s3"},
		},
		{
			name: "Multiple resources",
			code: `
				const vpc = new awsx.ec2.Vpc("custom");
				const cluster = new aws.ecs.Cluster("cluster");
				const db = new aws.rds.Instance("db");
			`,
			want: []string{"ec2", "ecs", "rds"}, // ec2 from awsx.ec2 (heuristic matches aws.ec2)
		},
		{
			name: "No resources",
			code: `const config = new pulumi.Config();`,
			want: nil,
		},
		{
			name: "Case insensitive",
			code: `new AWS.S3.Bucket`,
			want: []string{"s3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.extractResourceTypes(tt.code)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractResourceTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}
