package cmd

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var assetsPath string
var outputPath string
var force bool

func init() {
	spritesCmd.Flags().StringVarP(&assetsPath, "assets", "a", "", "path to assets directory")
	spritesCmd.Flags().StringVarP(&outputPath, "output", "o", "", "path to output file")
	spritesCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite output file if it exists")

	err := spritesCmd.MarkFlagRequired("assets")
	if err != nil {
		fmt.Println("Error marking flag required:", err)
	}

	err = spritesCmd.MarkFlagRequired("output")
	if err != nil {
		fmt.Println("Error marking flag required:", err)
	}

	rootCmd.AddCommand(spritesCmd)
}

var spritesCmd = &cobra.Command{
	Use:   "sprites [OPTIONS]",
	Short: "Generate code from sprites directory",
	Long:  `Output JavaScript file containing definitions for colors and sprites based on multiple images in assets directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(filepath.Dir(outputPath)); err != nil {
			log.Error("Directory of output path does not exist.")
			return
		}

		if _, err := os.Stat(outputPath); err == nil {
			if !force {
				log.Warn("Output file already exists. Please remove it first, or add --force flag to overwrite it.")
				return
			}
		}

		if _, err := os.Stat(assetsPath); err != nil {
			log.Error("Assets directory does not exist.")
			return
		}

		files, err := os.ReadDir(assetsPath)
		if err != nil {
			log.Error("Error reading assets directory: " + err.Error())
			return
		}

		pngs := make([]string, 0)

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".png") {
				pngs = append(pngs, file.Name())
				continue
			}

			if file.IsDir() {
				log.Warn("Assets directory contains a directory: " + file.Name())
				continue
			}

			log.Info("Skipping non-PNG file: " + file.Name())
		}

		if len(pngs) == 0 {
			log.Error("No PNG files found in assets directory.")
			return
		}

		type ColorMetadata struct {
			Color string
			Count int
			Files []string
			Index int
		}

		type SpriteMetadata struct {
			Rows [][]string
		}

		sprites := map[string]SpriteMetadata{}
		colors := map[string]ColorMetadata{}
		currentColorIndex := 0
		maxWidth := 0
		maxHeight := 0
		warnedAboutWidth := false
		warnedAboutHeight := false

		for _, png := range pngs {
			spriteName := strings.TrimSuffix(png, ".png")
			file, err := os.Open(filepath.Join(assetsPath, png))
			if err != nil {
				log.Error("Error opening file " + png + ": " + err.Error())
				continue
			}
			defer func() {
				if err := file.Close(); err != nil && err == nil {
					panic(err)
				}
			}()

			img, _, err := image.Decode(file)
			if err != nil {
				log.Error("Error decoding image: " + err.Error())
				return
			}

			bounds := img.Bounds()

			if bounds.Dx() > maxWidth {
				if maxWidth != 0 && !warnedAboutWidth {
					log.Warn("Images have different widths, which usually indicate a sprite sheet problem.")
					warnedAboutWidth = true
				}
				maxWidth = bounds.Dx()
			}

			if bounds.Dy() > maxHeight {
				if maxHeight != 0 && !warnedAboutHeight {
					log.Warn("Images have different heights, which usually indicate a sprite sheet problem.")
					warnedAboutHeight = true
				}

				maxHeight = bounds.Dy()
			}

			if _, exists := sprites[spriteName]; !exists {
				sprites[spriteName] = SpriteMetadata{
					Rows: make([][]string, bounds.Max.Y),
				}
			}

			rowI := 0
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				spriteRow := make([]string, bounds.Max.X)

				columnI := 0
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					c := img.At(x, y)
					r, g, b, a := c.RGBA()

					// RGBA() returns values in the range [0, 65535], scale them to [0, 255]
					r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)

					hexCodeRGBA := fmt.Sprintf("#%02x%02x%02x%02x", r8, g8, b8, a8)

					if hexCodeRGBA != "#00000000" {
						if _, exists := colors[hexCodeRGBA]; !exists {
							colors[hexCodeRGBA] = ColorMetadata{
								Color: hexCodeRGBA,
								Count: 1,
								Files: []string{png},
								Index: currentColorIndex,
							}
							currentColorIndex++
						} else {
							color := colors[hexCodeRGBA]
							color.Count++
	
							// check if file already present
							filePresent := false
							for _, file := range color.Files {
								if file == png {
									filePresent = true
									break
								}
							}
	
							if !filePresent {
								color.Files = append(colors[hexCodeRGBA].Files, png)
							}
	
							colors[hexCodeRGBA] = color
						}
					}
					
					
					colorIndexOfPixel := "."
					if hexCodeRGBA != "#00000000" {
						if colors[hexCodeRGBA].Index < 10 {
							colorIndexOfPixel = strconv.Itoa(colors[hexCodeRGBA].Index)
						} else {
							newIndex := colors[hexCodeRGBA].Index - 10
							charsMap := strings.Split("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", "")
							
							if newIndex - 1 > len(charsMap) {
								log.Error("Too many colors. You can only use up to 62 colors")
							}
							
							colorIndexOfPixel = charsMap[newIndex]
						}
					}
					
					spriteRow[columnI] = colorIndexOfPixel
					columnI++
				}

				sprites[spriteName].Rows[rowI] = spriteRow
				rowI++
			}
		}

		if len(colors) == 0 {
			log.Error("No colors found in PNG images.")
			return
		}

		if len(sprites) == 0 {
			log.Error("No sprites made from PNG images.")
			return
		}

		for _, color := range colors {
			word := "file"
			if color.Count > 1 {
				word = "files"
			}
			log.Debug(color.Color + " found " + strconv.Itoa(color.Count) + " times in " + strconv.Itoa(len(color.Files)) + " " + word)
		}
		
		log.Info(strconv.Itoa(len(colors)) + " colors found across all sprites")
		log.Info(strconv.Itoa(len(sprites)) + " sprites found across all PNG files")

		codeColors := make([]string, len(colors))
		{
			for hex, color := range colors {
				codeColors[color.Index] = hex
			}
		}

		codeSprites := make([]string, len(sprites))
		{
			i := 0
			for spriteName, sprite := range sprites {
				rows := sprite.Rows

				codeRows := make([]string, len(rows))
				for rowI, row := range rows {
					codeRows[rowI] = `			` + strings.Join(row, "")
				}

				codeSprites[i] = `		"` + spriteName + `": ` + "`" + `
` + strings.Join(codeRows, "\n") + `
		` + "`,"
				i++
			}
		}

		// TODO: This is ugly, use some templating engine. Your future self will thank you a lot.
		code := `var gameConfig = {
	cellWidth: ` + strconv.Itoa(maxWidth) + `,
	cellHeight: ` + strconv.Itoa(maxHeight) + `,
	colors: [
		"` + strings.Join(codeColors, `",
		"`) + `",
	],
	sprites: {
` + strings.Join(codeSprites, "\n") + `
	}
};`

		// wrtie code to outputPAth
		if outputPath != "" {
			err := os.WriteFile(outputPath, []byte(code), 0644)
			if err != nil {
				log.Errorf("Failed to write code to output file: %v", err)
			}
		}

		log.Logf(2, "Sprites configuration generated successfully.")
	},
}
