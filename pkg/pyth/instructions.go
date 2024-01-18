package pyth

import "encoding/binary"

const (
	Instruction_InitMapping = uint32(iota)
	Instruction_AddMapping
	Instruction_AddProduct
	Instruction_UpdProduct
	Instruction_AddPrice
	Instruction_AddPublisher
	Instruction_DelPublisher
	Instruction_UpdPrice
	Instruction_AggPrice
	Instruction_InitPrice
	Instruction_InitTest
	Instruction_UpdTest
	Instruction_SetMinPub
	Instruction_UpdPriceNoFailOnError
	instruction_count // number of different instruction types
)

func IsUpdatePriceInstruction(data []byte) bool {
	version := binary.LittleEndian.Uint32(data[0:4])
	ixNumber := binary.LittleEndian.Uint32(data[4:8])
	if version != Version {
		return false
	}

	return ixNumber == Instruction_UpdPrice || ixNumber == Instruction_UpdPriceNoFailOnError
}

func ParseUpdatePriceInstruction(data []byte) (price int64, conf uint64) {
	price = int64(binary.LittleEndian.Uint64(data[16:24]))
	conf = binary.LittleEndian.Uint64(data[24:32])
	return
}
