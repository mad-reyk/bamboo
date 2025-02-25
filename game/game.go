package game

import (
	"fmt"
	"github.com/3elDU/bamboo/blocks"
	"github.com/3elDU/bamboo/colors"
	"github.com/3elDU/bamboo/config"
	"github.com/3elDU/bamboo/event"
	"github.com/3elDU/bamboo/font"
	"github.com/3elDU/bamboo/game/inventory"
	"github.com/3elDU/bamboo/game/player"
	"github.com/3elDU/bamboo/game/widgets"
	"github.com/3elDU/bamboo/items"
	"github.com/3elDU/bamboo/scene_manager"
	"github.com/3elDU/bamboo/types"
	"github.com/3elDU/bamboo/widget"
	"github.com/3elDU/bamboo/world"
	"github.com/3elDU/bamboo/world_type"
	"github.com/MakeNowJust/heredoc"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"log"
)

type Game struct {
	widgets      *widget.Container
	debugWidgets *widget.Container

	paused    bool
	pauseMenu *pauseMenu

	world     *world.World
	player    *player.Player
	inventory *inventory.Inventory

	debugInfoVisible bool
}

func newGame(gameWorld *world.World, player *player.Player) *Game {
	game := &Game{
		widgets:      widget.NewWidgetContainer(),
		debugWidgets: widget.NewWidgetContainer(),

		pauseMenu: newPauseMenu(),

		world:     gameWorld,
		player:    player,
		inventory: inventory.NewInventory(),

		debugInfoVisible: false,
	}

	game.debugWidgets.AddTextWidget(
		"debug",
		&widgets.PerfWidget{Color: colors.Black},
	)

	return game
}

// Creates a game scene with a new world
func NewGameScene(metadata types.Save) *Game {
	w := world.NewWorld(metadata)
	game := newGame(
		w,
		player.NewPlayer(w),
	)
	game.player.SelectedWorld = metadata

	// perform a save immediately after the scene creation
	game.Save()

	return game
}

// Creates a game scene from existing world
func LoadGameScene(metadata types.Save) *Game {
	// load the player first, to determine which world to load
	loadedPlayer := player.LoadPlayer(metadata.BaseUUID)
	loadedWorld := world.Load(metadata.BaseUUID, loadedPlayer.SelectedWorld.UUID)
	return newGame(loadedWorld, loadedPlayer)
}

func (game *Game) Save() {
	game.world.Save()
	game.player.Save(game.world.Metadata())
}

func (game *Game) processInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		game.paused = !game.paused
		log.Printf("Escape pressed. Toggled pause menu. (%v)", game.paused)

		// trigger a world save when entering pause menu
		if game.paused {
			game.Save()
		}
	}

	if game.paused {
		switch game.pauseMenu.ButtonPressed() {
		case continueButtonPressed:
			game.paused = false
		case exitButtonPressed:
			game.Save()
			scene_manager.Pop()
		}
		return
	}

	game.player.Update(player.MovementVector{
		Left:  ebiten.IsKeyPressed(ebiten.KeyA),
		Right: ebiten.IsKeyPressed(ebiten.KeyD),
		Up:    ebiten.IsKeyPressed(ebiten.KeyW),
		Down:  ebiten.IsKeyPressed(ebiten.KeyS),
	}, game.world)

	// Check for key presses
	switch {
	// F3 toggles visibility of debug widgets
	case inpututil.IsKeyJustPressed(ebiten.KeyF3):
		game.debugInfoVisible = !game.debugInfoVisible
		log.Printf("Toggled visibility of debug info. (%v)", game.debugInfoVisible)

	// Interact with the nearby block
	case ebiten.IsKeyPressed(ebiten.KeyC):
		block := game.world.BlockAt(uint64(game.player.X), uint64(game.player.Y))
		drawable, ok := block.(types.DrawableBlock)
		if !ok {
			break
		}

		item := items.NewItemFromBlock(drawable)
		game.inventory.AddItem(item)

	// Use the item in hand
	case ebiten.IsKeyPressed(ebiten.KeyF):
		itemInHand := game.inventory.Slots[game.inventory.SelectedSlot].Item
		if itemInHand == nil {
			break
		}
		itemInHand.Use(game.world, types.Vec2u{
			X: uint64(game.player.X),
			Y: uint64(game.player.Y),
		})

	// Inventory slots selection
	case ebiten.IsKeyPressed(ebiten.KeyDigit1):
		game.inventory.SelectSlot(0)
	case ebiten.IsKeyPressed(ebiten.KeyDigit2):
		game.inventory.SelectSlot(1)
	case ebiten.IsKeyPressed(ebiten.KeyDigit3):
		game.inventory.SelectSlot(2)
	case ebiten.IsKeyPressed(ebiten.KeyDigit4):
		game.inventory.SelectSlot(3)
	case ebiten.IsKeyPressed(ebiten.KeyDigit5):
		game.inventory.SelectSlot(4)
	}

	_, yoff := ebiten.Wheel()
	if yoff < 0 {
		game.inventory.SelectSlot(game.inventory.SelectedSlot + 1)
	} else if yoff > 0 {
		game.inventory.SelectSlot(game.inventory.SelectedSlot - 1)
	}
}

func (game *Game) updateLogic() {
	if game.paused {
		return
	}

	game.world.Update()
	game.widgets.Update()
	if game.debugInfoVisible {
		game.debugWidgets.Update()
	}

	// perform autosave each N ticks
	if scene_manager.Ticks()%config.WorldAutosaveDelay == 0 {
		game.Save()
	}
}

func (game *Game) handleEvents() {
	for _, ev := range event.GetEvents() {
		switch ev.Type() {
		case event.CaveEntered:
			// save the previous world before switching to a new one
			game.Save()

			caveID := ev.Args().(event.CaveEnteredArgs).ID

			caveExit := blocks.NewCaveEntranceBlock(game.world.Metadata().UUID)

			metadata := types.Save{
				Name:      game.world.Metadata().Name,
				BaseUUID:  game.world.Metadata().BaseUUID,
				UUID:      caveID,
				Seed:      int64(caveID.ID()),
				WorldType: world_type.Cave,
			}

			var newWorld *world.World
			// Check if the world already exists on disk
			if world.ExistsOnDisk(metadata) {
				newWorld = world.Load(metadata.BaseUUID, metadata.UUID)
			} else {
				newWorld = world.NewWorld(metadata)
			}

			game.player = player.NewPlayer(newWorld)

			// if we're switching from cave to overworld, don't place the cave exit.
			// also don't place cave exit if that chunk already exists on disk, so we don't overwrite it
			if newWorld.Metadata().WorldType == world_type.Cave && !world.ChunkExistsOnDisk(
				newWorld.Metadata(),
				uint64(game.player.X+2)/16, uint64(game.player.Y)/16,
			) {
				// place a portal to overworld next to the player
				newWorld.SetBlock(uint64(game.player.X)+2, uint64(game.player.Y), caveExit)
			}

			game.world = newWorld

			game.Save()
		}
	}
}

func (game *Game) Update() {
	game.processInput()
	game.updateLogic()
	game.handleEvents()
}

func (game *Game) Draw(screen *ebiten.Image) {
	game.world.Render(screen, game.player.X, game.player.Y, config.UIScaling)
	game.player.Render(screen, config.UIScaling, game.paused)
	game.inventory.Render(screen)

	game.widgets.Render(screen)
	if game.debugInfoVisible {
		game.debugWidgets.Render(screen)
	}

	if game.debugInfoVisible {
		font.RenderFont(screen,
			fmt.Sprintf(
				heredoc.Doc(`
					player pos:		%.2f, %.2f
					world seed:		%v
					UI scaling:		%v
				`),
				game.player.X, game.player.Y, game.world.Seed(), config.UIScaling,
			),
			0, 0, colors.Black,
		)
	}

	// draw pause menu
	if game.paused {
		err := game.pauseMenu.Draw(screen)
		if err != nil {
			log.Panicf("error while rendering pause menu - %v", err)
		}
	}
}

func (game *Game) Destroy() {
	game.Save()
	log.Println("GameScene.Destroy() called")
}
