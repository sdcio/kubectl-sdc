package runningconfig

import (
	"context"
	"strings"
	"testing"

	v1alpha1 "github.com/sdcio/config-server/apis/config/v1alpha1"
	mockrunningconfig "github.com/sdcio/kubectl-sdc/mocks/runningconfig"
	"github.com/sdcio/kubectl-sdc/pkg/client"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestRunXPathDecodesPathValuesAndRequestsXPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := mockrunningconfig.NewMockRunningConfigClient(ctrl)

	// Build a PathValues proto matching what the server returns for ?format=xpath
	pathName, err := sdcpb.ParsePath("/system/name")
	if err != nil {
		t.Fatalf("ParsePath(/system/name) unexpected error: %v", err)
	}
	pathAdminState, err := sdcpb.ParsePath("/system/snmp/network-instance[name=mgmt]/admin-state")
	if err != nil {
		t.Fatalf("ParsePath(/system/snmp/...) unexpected error: %v", err)
	}
	pv := &sdcpb.PathValues{
		PathValues: []*sdcpb.PathValue{
			{
				Path:  pathName,
				Value: &sdcpb.TypedValue{Value: &sdcpb.TypedValue_StringVal{StringVal: "srl1"}},
			},
			{
				Path:  pathAdminState,
				Value: &sdcpb.TypedValue{Value: &sdcpb.TypedValue_StringVal{StringVal: "enable"}},
			},
		},
	}
	raw, err := protojson.Marshal(pv)
	if err != nil {
		t.Fatalf("protojson.Marshal() unexpected error: %v", err)
	}

	cl.EXPECT().GetRunningConfig(gomock.Any(), "default", "srl1", client.FormatXPath).
		Return(&v1alpha1.TargetRunningConfig{Value: string(raw)}, nil)

	got, err := Run(context.Background(), cl, "default", "srl1", client.FormatXPath)
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("Run() returned empty output")
	}
	// Output must be sorted path: value lines, one per entry
	for _, want := range []string{"system/name: srl1", "admin-state: enable"} {
		if !strings.Contains(got, want) {
			t.Errorf("Run() output missing %q\ngot:\n%s", want, got)
		}
	}
}
