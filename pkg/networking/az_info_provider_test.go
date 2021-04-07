package networking

import (
	ec2sdk "github.com/aws/aws-sdk-go/service/ec2"
	"reflect"
	"testing"
)

func Test_computeAZIDsWithoutAZInfo(t *testing.T) {
	type args struct {
		availabilityZoneIDs []string
		azInfoByAZID        map[string]ec2sdk.AvailabilityZone
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := computeAZIDsWithoutAZInfo(tt.args.availabilityZoneIDs, tt.args.azInfoByAZID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("computeAZIDsWithoutAZInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
