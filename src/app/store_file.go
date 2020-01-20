package app

import (
	"io/ioutil"
	"path"
)

func (a *App) SaveToLogFile(name string, data []byte) error {
	return ioutil.WriteFile(path.Join(a.config.LogHttpApiDir, name), data, 0644)
}
