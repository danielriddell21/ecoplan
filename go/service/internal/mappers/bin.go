package mappers

import pb "github.com/ecoscan/service/proto"

func ToBinColour(bin string) pb.RecyclingItem_BinColour {
	switch bin {
	case "blue":
		return pb.RecyclingItem_BIN_COLOUR_BLUE
	case "green":
		return pb.RecyclingItem_BIN_COLOUR_GREEN
	case "brown":
		return pb.RecyclingItem_BIN_COLOUR_BROWN
	case "black":
		return pb.RecyclingItem_BIN_COLOUR_BLACK
	default:
		return pb.RecyclingItem_BIN_COLOUR_UNSPECIFIED
	}
}

func ToBinType(bin string) pb.RecyclingItem_BinType {
	switch bin {
	case "blue":
		return pb.RecyclingItem_BIN_TYPE_GLASS
	case "green":
		return pb.RecyclingItem_BIN_TYPE_RECYCLING
	case "brown":
		return pb.RecyclingItem_BIN_TYPE_PAPER
	case "black":
		return pb.RecyclingItem_BIN_TYPE_WASTE
	default:
		return pb.RecyclingItem_BIN_TYPE_UNSPECIFIED
	}
}
