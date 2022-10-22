/*
Scene is a distinct state of the program, that displays specific state of the game.
For example: main menu scene, "new game" scene, playing scene, death scene, etc.
*/
package scene

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/exp/slices"
)

type Scene interface {
	Update(manager *SceneManager) error
	Draw(screen *ebiten.Image)

	// called when the scene is about to be deleted
	Destroy()
}

type SceneManager struct {
	currentScene Scene
	queue        []Scene
}

func InitSceneManager() *SceneManager {
	ebiten.SetWindowClosingHandled(true)
	return &SceneManager{
		currentScene: nil,
		queue:        make([]Scene, 0),
	}
}

// Must be called from Scene.Update()
// Exits current scene, and switches to next in the queue
// If the queue is empty, exits
func (manager *SceneManager) End() {
	if manager.currentScene != nil {
		manager.currentScene.Destroy()
	}

	if len(manager.queue) != 0 {
		next := manager.queue[0]
		manager.currentScene = next

		// delete scene from the queue
		manager.queue[0] = nil
		manager.queue = slices.Delete(manager.queue, 0, 1)
	} else {
		manager.currentScene = nil
	}
}

// Switches to the given scene, disregarding the queue
func (manager *SceneManager) Switch(next Scene) {

}

// Pushes scene to the end of the queue
func (manager *SceneManager) Push(sc Scene) {
	manager.queue = append(manager.queue, sc)
}

func (manager *SceneManager) Update() error {
	if ebiten.IsWindowBeingClosed() {
		log.Println("SceneManager.Update() - Handling window close")

		if manager.currentScene != nil {
			manager.currentScene.Destroy()
		}
		for _, sc := range manager.queue {
			sc.Destroy()
		}

		return fmt.Errorf("exit")
	}

	if manager.currentScene == nil {
		if len(manager.queue) != 0 {
			manager.End()
		} else {
			log.Println("SceneManager.Update() - No scenes left to display. Exiting!")
			return fmt.Errorf("exit")
		}
	}

	if err := manager.currentScene.Update(manager); err != nil {
		return err
	}

	return nil
}

func (manager *SceneManager) Draw(screen *ebiten.Image) {
	manager.currentScene.Draw(screen)
}

func (manager *SceneManager) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
