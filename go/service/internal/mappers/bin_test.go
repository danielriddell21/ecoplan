package mappers_test

import (
	"testing"

	"github.com/ecoscan/service/internal/mappers"
	pb "github.com/ecoscan/service/proto"
)

func TestToBinColour(t *testing.T) {
	cases := []struct {
		input string
		want  pb.RecyclingItem_BinColour
	}{
		{"blue", pb.RecyclingItem_BIN_COLOUR_BLUE},
		{"green", pb.RecyclingItem_BIN_COLOUR_GREEN},
		{"brown", pb.RecyclingItem_BIN_COLOUR_BROWN},
		{"black", pb.RecyclingItem_BIN_COLOUR_BLACK},
		{"unknown", pb.RecyclingItem_BIN_COLOUR_UNSPECIFIED},
		{"", pb.RecyclingItem_BIN_COLOUR_UNSPECIFIED},
	}
	for _, tc := range cases {
		got := mappers.ToBinColour(tc.input)
		if got != tc.want {
			t.Errorf("ToBinColour(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestToBinType(t *testing.T) {
	cases := []struct {
		input string
		want  pb.RecyclingItem_BinType
	}{
		{"blue", pb.RecyclingItem_BIN_TYPE_GLASS},
		{"green", pb.RecyclingItem_BIN_TYPE_RECYCLING},
		{"brown", pb.RecyclingItem_BIN_TYPE_PAPER},
		{"black", pb.RecyclingItem_BIN_TYPE_WASTE},
		{"unknown", pb.RecyclingItem_BIN_TYPE_UNSPECIFIED},
		{"", pb.RecyclingItem_BIN_TYPE_UNSPECIFIED},
	}
	for _, tc := range cases {
		got := mappers.ToBinType(tc.input)
		if got != tc.want {
			t.Errorf("ToBinType(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
