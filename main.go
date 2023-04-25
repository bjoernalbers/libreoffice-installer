// libreoffice-installer
package main

import (
	"fmt"
	"os"
)

func main() {
	app := App{"/Applications/LibreOffice.app"}
	if app.IsMissing() {
		fmt.Println("LibreOffice is missing. Installation required.")
	}
}

type App struct {
	Path string
}

func (a *App) IsMissing() bool {
	_, err := os.Stat(a.Path)
	if err == nil {
		return false
	}
	return true
}
