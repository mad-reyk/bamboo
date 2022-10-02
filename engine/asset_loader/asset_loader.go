/*
	Loads all assets from the folder
*/

package asset_loader

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/3elDU/bamboo/config"
	"github.com/3elDU/bamboo/engine"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type AssetList struct {
	Fonts    map[string]*ttf.Font
	Textures map[string]*sdl.Texture
}

// removes file extension, and other parts from the filename
// Example: assets/pictures/picture.png -> picture
func cleanPath(path string) string {
	return strings.Replace(filepath.Base(path), filepath.Ext(path), "", 1)
}

func LoadAssets(engine *engine.Engine, dir string) *AssetList {
	assetList := &AssetList{
		Fonts:    make(map[string]*ttf.Font),
		Textures: make(map[string]*sdl.Texture),
	}

	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			switch filepath.Ext(path) {
			case ".png":
				surf, err := img.Load(path)
				if err != nil {
					return err
				}
				tex, err := engine.Ren.CreateTextureFromSurface(surf)
				if err != nil {
					return err
				}
				assetList.Textures[cleanPath(path)] = tex
			case ".ttf":
				font, err := ttf.OpenFont(path, config.FONT_SIZE)
				if err != nil {
					return err
				}
				assetList.Fonts[cleanPath(path)] = font
			}
		}
		return nil
	})

	return assetList
}
