package zapp_test

import (
	"os"
	"testing"
	"time"

	zapp "github.com/srikrsna/zapproto"
	testspb "github.com/srikrsna/zapproto/test_data"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestP(t *testing.T) {
	dob, _ := time.Parse(time.RFC3339, "2012-01-02T15:04:05Z07:00")
	md, err := structpb.NewStruct(map[string]interface{}{
		"string": "StringValue",
		"number": 123.92,
		"list": []interface{}{
			map[string]interface{}{
				"inner": "InnerValue",
			},
		},
		"null": nil,
		"bool": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	p := &testspb.Person{
		Name:          "Hero",
		Age:           32,
		Dob:           timestamppb.New(dob),
		MaritalStatus: testspb.MaritalStatus_MaritalStatus_SINGLE,
		Alive:         false,
		FavouriteBooks: map[string]*testspb.Books{
			"MarkTwain": {Books: []string{"Adventours of huckle berry finn"}},
		},
		FavDays: []*timestamppb.Timestamp{
			timestamppb.New(dob),
			timestamppb.New(dob),
			timestamppb.New(dob),
		},
		Metadata: md,
		Siblings: []*testspb.Person{
			{
				Name:          "Sister",
				Age:           28,
				Dob:           timestamppb.New(dob.AddDate(-4, 0, 0)),
				MaritalStatus: testspb.MaritalStatus_MaritalStatus_UNSPECIFIED,
				HoursPerWeek:  durationpb.New(40 * time.Hour),
				Alive:         true,
				Gender: &testspb.Person_Female{
					Female: true,
				},
			},
		},
	}

	f, err := os.Create("test_data/golden.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	l := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zapcore.EncoderConfig{}), zapcore.AddSync(f), zapcore.InfoLevel))
	defer l.Sync()

	l.Info("Sample", zapp.P("person", p))
}

