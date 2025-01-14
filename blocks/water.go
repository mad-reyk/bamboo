package blocks

import (
	"encoding/gob"

	"github.com/3elDU/bamboo/asset_loader"
	"github.com/3elDU/bamboo/types"
)

func init() {
	gob.Register(WaterState{})
}

type WaterState struct {
	ConnectedBlockState
	CollidableBlockState
}

type WaterBlock struct {
	connectedBlock
	collidableBlock
}

func NewWaterBlock() *WaterBlock {
	return &WaterBlock{
		connectedBlock: connectedBlock{
			baseBlock: baseBlock{
				blockType: Water,
			},
			tex:        asset_loader.ConnectedTexture("lake", false, false, false, false),
			connectsTo: []types.BlockType{Water},
		},
		collidableBlock: collidableBlock{
			collidable:  false,
			playerSpeed: 0.2,
		},
	}
}

func (b *WaterBlock) State() interface{} {
	return WaterState{
		ConnectedBlockState:  b.connectedBlock.State().(ConnectedBlockState),
		CollidableBlockState: b.collidableBlock.State().(CollidableBlockState),
	}
}

func (b *WaterBlock) LoadState(s interface{}) {
	state := s.(WaterState)
	b.connectedBlock.LoadState(state.ConnectedBlockState)
	b.collidableBlock.LoadState(state.CollidableBlockState)
}
