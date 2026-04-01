package mappers

import pb "github.com/ecoscan/service/proto"

func ToBinColour(bin string) pb.RecyclingItem_BinColour {
	switch bin {
	case "blue":
		return pb.RecyclingItem_BLUE
	case "green":
		return pb.RecyclingItem_GREEN
	case "brown":
		return pb.RecyclingItem_BROWN
	case "black":
		return pb.RecyclingItem_BLACK
	default:
		return pb.RecyclingItem_BIN_COLOUR_UNKNOWN
	}
}

func ToBinType(bin string) pb.RecyclingItem_BinType {
	switch bin {
	case "blue":
		return pb.RecyclingItem_GLASS
	case "green":
		return pb.RecyclingItem_RECYCLING
	case "brown":
		return pb.RecyclingItem_PAPER
	case "black":
		return pb.RecyclingItem_WASTE
	default:
		return pb.RecyclingItem_BIN_TYPE_UNKNOWN
	}
}
